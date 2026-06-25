package maigo

import (
	"context"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	pgrpc "github.com/tikhonp/maigo/internal/grpc/recordspb"
	pjson "github.com/tikhonp/maigo/internal/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

//go:embed grpc_ca.pem
var grpcCACert []byte

const (
	grpcPort         = 50051
	grpcServerName   = "medsenger.ru" // matches the cert SAN; mirrors Python's ssl_target_name_override.
	grpcDefaultHost  = "medsenger.ru"
	grpcMaxMsgSize   = 50 * 1024 * 1024
	grpcCallTimeout  = 20 * time.Second
	grpcRequestTries = 3
)

// errGRPCNotFound signals that a gRPC call returned codes.NotFound, i.e. the
// resource legitimately does not exist. Callers treat it as an empty result
// rather than falling back to REST (mirroring the Python client).
var errGRPCNotFound = errors.New("maigo: grpc resource not found")

// grpcClient wraps the generated Records gRPC client. It manages a lazily
// established TLS connection, per-call auth + retries, and a category cache,
// mirroring the Python RecordsClient.
type grpcClient struct {
	apiKey string
	host   string

	mu   sync.Mutex
	conn *grpc.ClientConn
	stub pgrpc.RecordsClient

	catMu             sync.Mutex
	categoriesByID    map[int64]*pgrpc.Category
	categoryIDsByName map[string][]int64
}

func newGRPCClient(apiKey, host string) *grpcClient {
	if host == "" {
		host = grpcDefaultHost
	}
	return &grpcClient{apiKey: apiKey, host: host}
}

func (g *grpcClient) connectLocked() error {
	if g.conn != nil {
		_ = g.conn.Close()
		g.conn = nil
		g.stub = nil
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(grpcCACert) {
		return errors.New("maigo: failed to parse embedded gRPC CA certificate")
	}
	creds := credentials.NewClientTLSFromCert(pool, grpcServerName)
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", g.host, grpcPort),
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(grpcMaxMsgSize),
			grpc.MaxCallSendMsgSize(grpcMaxMsgSize),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return err
	}
	g.conn = conn
	g.stub = pgrpc.NewRecordsClient(conn)
	return nil
}

func (g *grpcClient) ensureStub() (pgrpc.RecordsClient, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.stub == nil {
		if err := g.connectLocked(); err != nil {
			return nil, err
		}
	}
	return g.stub, nil
}

func (g *grpcClient) reconnect() {
	g.mu.Lock()
	defer g.mu.Unlock()
	_ = g.connectLocked()
}

func (g *grpcClient) close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.conn != nil {
		err := g.conn.Close()
		g.conn = nil
		g.stub = nil
		return err
	}
	return nil
}

// grpcCall runs fn with retries and per-call auth metadata, reconnecting between
// attempts. A codes.NotFound result is converted to errGRPCNotFound.
func grpcCall[T any](g *grpcClient, fn func(ctx context.Context, stub pgrpc.RecordsClient) (T, error)) (T, error) {
	var zero T
	var lastErr error
	for range grpcRequestTries {
		stub, err := g.ensureStub()
		if err != nil {
			lastErr = err
			g.reconnect()
			continue
		}
		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", g.apiKey)
		ctx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
		out, err := fn(ctx, stub)
		cancel()
		if err == nil {
			return out, nil
		}
		if status.Code(err) == codes.NotFound {
			return zero, errGRPCNotFound
		}
		lastErr = err
		g.reconnect()
	}
	return zero, lastErr
}

// --- categories ---

func (g *grpcClient) getCategories() ([]Category, error) {
	resp, err := grpcCall(g, func(ctx context.Context, stub pgrpc.RecordsClient) (*pgrpc.CategoryList, error) {
		return stub.GetCategoryList(ctx, &pgrpc.Empty{})
	})
	if err != nil {
		return nil, err
	}
	return g.cacheAndPresentCategories(resp.GetCategories()), nil
}

func (g *grpcClient) getCategoriesForUser(userID int) ([]Category, error) {
	resp, err := grpcCall(g, func(ctx context.Context, stub pgrpc.RecordsClient) (*pgrpc.CategoryList, error) {
		return stub.GetCategoryListForUser(ctx, &pgrpc.User{Id: int32(userID)})
	})
	if err != nil {
		return nil, err
	}
	return g.cacheAndPresentCategories(resp.GetCategories()), nil
}

func (g *grpcClient) cacheAndPresentCategories(cats []*pgrpc.Category) []Category {
	g.catMu.Lock()
	defer g.catMu.Unlock()
	if g.categoriesByID == nil {
		g.categoriesByID = make(map[int64]*pgrpc.Category)
		g.categoryIDsByName = make(map[string][]int64)
	}
	out := make([]Category, 0, len(cats))
	for _, c := range cats {
		g.cacheCategoryLocked(c)
		out = append(out, categoryFromPB(c))
	}
	return out
}

func (g *grpcClient) cacheCategoryLocked(c *pgrpc.Category) {
	g.categoriesByID[c.GetId()] = c
	name := c.GetName()
	for _, id := range g.categoryIDsByName[name] {
		if id == c.GetId() {
			return
		}
	}
	g.categoryIDsByName[name] = append(g.categoryIDsByName[name], c.GetId())
}

func (g *grpcClient) ensureCategories() error {
	g.catMu.Lock()
	empty := len(g.categoriesByID) == 0
	g.catMu.Unlock()
	if empty {
		if _, err := g.getCategories(); err != nil {
			return err
		}
	}
	return nil
}

func (g *grpcClient) categoryIDsForNames(names []string) []int64 {
	g.catMu.Lock()
	defer g.catMu.Unlock()
	seen := make(map[int64]bool)
	var ids []int64
	for _, n := range names {
		for _, id := range g.categoryIDsByName[n] {
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func (g *grpcClient) lookupCategory(id int64) *Category {
	g.catMu.Lock()
	pb, ok := g.categoriesByID[id]
	g.catMu.Unlock()
	if !ok {
		if err := g.ensureCategories(); err == nil {
			g.catMu.Lock()
			pb, ok = g.categoriesByID[id]
			g.catMu.Unlock()
		}
	}
	if !ok {
		return nil
	}
	c := categoryFromPB(pb)
	return &c
}

// --- records ---

func (g *grpcClient) getRecordByID(recordID int) (*MedicalRecord, error) {
	resp, err := grpcCall(g, func(ctx context.Context, stub pgrpc.RecordsClient) (*pgrpc.Record, error) {
		return stub.GetRecordById(ctx, &pgrpc.RecordRequest{Id: int64(recordID)})
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errGRPCNotFound
	}
	rec := g.recordFromPB(resp)
	return &rec, nil
}

func (g *grpcClient) getRecords(q *pgrpc.RecordQuery) ([]MedicalRecord, error) {
	resp, err := grpcCall(g, func(ctx context.Context, stub pgrpc.RecordsClient) (*pgrpc.RecordList, error) {
		return stub.GetRecords(ctx, q)
	})
	if err != nil {
		if errors.Is(err, errGRPCNotFound) {
			return []MedicalRecord{}, nil
		}
		return nil, err
	}
	return g.presentRecords(resp.GetRecords()), nil
}

func (g *grpcClient) countRecords(q *pgrpc.RecordQuery) (int, error) {
	resp, err := grpcCall(g, func(ctx context.Context, stub pgrpc.RecordsClient) (*pgrpc.RecordList, error) {
		return stub.CountRecords(ctx, q)
	})
	if err != nil {
		if errors.Is(err, errGRPCNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return int(resp.GetCount()), nil
}

func (g *grpcClient) getMultipleRecords(queries []*pgrpc.RecordQuery) ([][]MedicalRecord, error) {
	resp, err := grpcCall(g, func(ctx context.Context, stub pgrpc.RecordsClient) (*pgrpc.MultiRecordAnswer, error) {
		return stub.GetMultipleRecords(ctx, &pgrpc.MultiRecordQuery{Queries: queries})
	})
	if err != nil {
		return nil, err
	}
	answers := resp.GetAnswers()
	out := make([][]MedicalRecord, len(answers))
	for i, ans := range answers {
		out[i] = g.presentRecords(ans.GetRecords())
	}
	return out, nil
}

func (g *grpcClient) presentRecords(records []*pgrpc.Record) []MedicalRecord {
	out := make([]MedicalRecord, 0, len(records))
	for _, r := range records {
		out = append(out, g.recordFromPB(r))
	}
	return out
}

func (g *grpcClient) recordFromPB(r *pgrpc.Record) MedicalRecord {
	cat := g.lookupCategory(r.GetCategoryId())
	rec := MedicalRecord{
		ID:        int(r.GetId()),
		Value:     convertValue(r.GetValue(), cat),
		Additions: parseJSONArray(r.GetAdditions()),
		Params:    parseJSONAny(r.GetParams()),
		Group:     r.GetGroup(),
		Time:      time.Unix(r.GetCreatedAt(), 0),
		Uploaded:  time.Unix(r.GetUpdatedAt(), 0),
	}
	if src := r.GetSource(); src != nil {
		rec.Source = MedicalRecordSource{ID: int(src.GetId()), Name: src.GetName()}
	}
	for _, f := range r.GetAttachedFiles() {
		rec.AttachedFiles = append(rec.AttachedFiles, MedicalRecordFile{
			ID:   int(f.GetId()),
			Name: f.GetName(),
			Type: f.GetType(),
		})
	}
	if cat != nil {
		rec.Category = *cat
	}
	return rec
}

// --- presentation helpers ---

func categoryFromPB(c *pgrpc.Category) Category {
	return Category{
		ID:                    int(c.GetId()),
		Name:                  c.GetName(),
		Description:           c.GetDescription(),
		Unit:                  c.GetUnit(),
		Type:                  c.GetType(),
		DefaultRepresentation: c.GetDefaultRepresentation(),
		IsLegacy:              c.GetIsLegacy(),
		Subcategory:           c.GetSubcategory(),
		DoctorCanAdd:          c.GetDoctorCanAdd(),
		DoctorCanReplace:      c.GetDoctorCanReplace(),
	}
}

// convertValue mirrors the Python client's value coercion: integer/float
// categories return numbers, everything else stays a string.
func convertValue(raw string, cat *Category) any {
	if cat != nil {
		switch cat.Type {
		case "integer":
			if f, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64); err == nil {
				return int(f)
			}
		case "float":
			if f, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64); err == nil {
				return math.Round(f*1000) / 1000
			}
		}
	}
	return raw
}

func parseJSONArray(s string) []any {
	if s == "" {
		return nil
	}
	var v []any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil
	}
	return v
}

func parseJSONAny(s string) any {
	if s == "" {
		return nil
	}
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil
	}
	return v
}

func splitCategoryNames(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func tsToUnixPtr(t *pjson.Timestamp) *int64 {
	if t == nil {
		return nil
	}
	v := t.Unix()
	return &v
}

func optInt64(v int) *int64 {
	if v == 0 {
		return nil
	}
	x := int64(v)
	return &x
}

// --- Client-level gRPC bridge ---

// MultiQuery describes one query in a GetMultipleRecords batch.
type MultiQuery struct {
	ContractID int
	Options    []GetRecordsOption
}

// buildRecordQuery resolves a contract to a gRPC user_id and translates record
// options into a RecordQuery. The bool is false when category names were given
// but none resolved to known categories (i.e. the result is known-empty).
func (c *Client) buildRecordQuery(contractID int, o *getRecordsOptions) (*pgrpc.RecordQuery, bool, error) {
	uid, err := c.resolveUserID(contractID)
	if err != nil {
		return nil, false, err
	}
	q := &pgrpc.RecordQuery{
		UserId:        int32(uid),
		WithGroup:     o.SameGroup,
		FromTimestamp: tsToUnixPtr(o.From),
		ToTimestamp:   tsToUnixPtr(o.To),
		Offset:        optInt64(o.Offset),
		Limit:         optInt64(o.Limit),
	}
	if names := splitCategoryNames(o.CategoryName); len(names) > 0 {
		if err := c.grpc.ensureCategories(); err != nil {
			return nil, false, err
		}
		ids := c.grpc.categoryIDsForNames(names)
		if len(ids) == 0 {
			return nil, false, nil
		}
		q.CategoryIds = ids
	}
	return q, true, nil
}

func (c *Client) grpcGetRecords(contractID int, o *getRecordsOptions) ([]MedicalRecord, error) {
	q, ok, err := c.buildRecordQuery(contractID, o)
	if err != nil {
		return nil, err
	}
	if !ok {
		return []MedicalRecord{}, nil
	}
	return c.grpc.getRecords(q)
}

func (c *Client) grpcGetMultipleRecords(queries []MultiQuery) ([][]MedicalRecord, error) {
	pbQueries := make([]*pgrpc.RecordQuery, 0, len(queries))
	knownEmpty := make(map[int]bool)
	for i, mq := range queries {
		o := getRecordsOptions{TokenAndContractRequest: c.tokenAndContractRequest(mq.ContractID)}
		applyGetRecordsOptions(&o, mq.Options...)
		q, ok, err := c.buildRecordQuery(mq.ContractID, &o)
		if err != nil {
			return nil, err
		}
		if !ok {
			knownEmpty[i] = true
			continue
		}
		pbQueries = append(pbQueries, q)
	}
	answers, err := c.grpc.getMultipleRecords(pbQueries)
	if err != nil {
		return nil, err
	}
	// Re-align answers (one per sent query) back to the original positions.
	out := make([][]MedicalRecord, len(queries))
	ai := 0
	for i := range queries {
		if knownEmpty[i] {
			out[i] = []MedicalRecord{}
			continue
		}
		if ai < len(answers) {
			out[i] = answers[ai]
			ai++
		} else {
			out[i] = []MedicalRecord{}
		}
	}
	return out, nil
}

// GetMultipleRecords fetches records for several queries in one batch. With gRPC
// enabled it issues a single GetMultipleRecords call; otherwise (or on gRPC
// error) it falls back to running each query through GetRecords individually.
func (c *Client) GetMultipleRecords(queries []MultiQuery) ([][]MedicalRecord, error) {
	if c.grpc != nil {
		if result, err := c.grpcGetMultipleRecords(queries); err == nil {
			return result, nil
		}
	}
	out := make([][]MedicalRecord, len(queries))
	for i, q := range queries {
		recs, err := c.GetRecords(q.ContractID, q.Options...)
		if err != nil {
			return nil, err
		}
		out[i] = recs
	}
	return out, nil
}

// CountRecords returns how many records match the query. With gRPC enabled it
// uses the dedicated CountRecords RPC; otherwise it falls back to counting the
// records returned by GetRecords.
func (c *Client) CountRecords(contractID int, opts ...GetRecordsOption) (int, error) {
	if c.grpc != nil {
		o := getRecordsOptions{TokenAndContractRequest: c.tokenAndContractRequest(contractID)}
		applyGetRecordsOptions(&o, opts...)
		if q, ok, err := c.buildRecordQuery(contractID, &o); err == nil {
			if !ok {
				return 0, nil
			}
			if n, err := c.grpc.countRecords(q); err == nil {
				return n, nil
			}
		}
	}
	recs, err := c.GetRecords(contractID, opts...)
	if err != nil {
		return 0, err
	}
	return len(recs), nil
}

// InitOption configures a Client at construction time (see Init).
type InitOption interface {
	apply(*Client)
}

type funcInitOption struct {
	f func(*Client)
}

func (o *funcInitOption) apply(c *Client) {
	o.f(c)
}

func newFuncInitOption(f func(*Client)) *funcInitOption {
	return &funcInitOption{f: f}
}

// WithGRPC enables the gRPC transport for record and category reads, with
// automatic fallback to REST on any gRPC error. host is the gRPC endpoint; pass
// "" to use the default (medsenger.ru). The connection is established lazily on
// first use; call Client.Close when done.
func WithGRPC(host string) InitOption {
	return newFuncInitOption(func(c *Client) {
		if host == "" {
			host = grpcDefaultHost
		}
		c.grpcHost = host
	})
}

// resolveUserID maps a contract id to the gRPC user_id, caching the result. It
// uses the same value the Python client does: the "id" field of patient info.
func (c *Client) resolveUserID(contractID int) (int, error) {
	c.userMu.Lock()
	if uid, ok := c.userCache[contractID]; ok {
		c.userMu.Unlock()
		return uid, nil
	}
	c.userMu.Unlock()

	info, err := c.GetContractInfo(contractID)
	if err != nil {
		return 0, err
	}
	c.userMu.Lock()
	if c.userCache == nil {
		c.userCache = make(map[int]int)
	}
	c.userCache[contractID] = info.ID
	c.userMu.Unlock()
	return info.ID, nil
}

// Close releases the gRPC connection, if any. It is safe to call on a Client
// that does not use gRPC.
func (c *Client) Close() error {
	if c.grpc != nil {
		return c.grpc.close()
	}
	return nil
}
