// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math"
	"rtglabs-go/ent/bodyweight"
	"rtglabs-go/ent/predicate"
	"rtglabs-go/ent/privatetoken"
	"rtglabs-go/ent/profile"
	"rtglabs-go/ent/session"
	"rtglabs-go/ent/user"
	"rtglabs-go/ent/workout"
	"rtglabs-go/ent/workoutlog"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserQuery is the builder for querying User entities.
type UserQuery struct {
	config
	ctx              *QueryContext
	order            []user.OrderOption
	inters           []Interceptor
	predicates       []predicate.User
	withBodyweights  *BodyweightQuery
	withSessions     *SessionQuery
	withProfile      *ProfileQuery
	withWorkouts     *WorkoutQuery
	withWorkoutLogs  *WorkoutLogQuery
	withPrivateToken *PrivateTokenQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the UserQuery builder.
func (uq *UserQuery) Where(ps ...predicate.User) *UserQuery {
	uq.predicates = append(uq.predicates, ps...)
	return uq
}

// Limit the number of records to be returned by this query.
func (uq *UserQuery) Limit(limit int) *UserQuery {
	uq.ctx.Limit = &limit
	return uq
}

// Offset to start from.
func (uq *UserQuery) Offset(offset int) *UserQuery {
	uq.ctx.Offset = &offset
	return uq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (uq *UserQuery) Unique(unique bool) *UserQuery {
	uq.ctx.Unique = &unique
	return uq
}

// Order specifies how the records should be ordered.
func (uq *UserQuery) Order(o ...user.OrderOption) *UserQuery {
	uq.order = append(uq.order, o...)
	return uq
}

// QueryBodyweights chains the current query on the "bodyweights" edge.
func (uq *UserQuery) QueryBodyweights() *BodyweightQuery {
	query := (&BodyweightClient{config: uq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := uq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(user.Table, user.FieldID, selector),
			sqlgraph.To(bodyweight.Table, bodyweight.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, user.BodyweightsTable, user.BodyweightsColumn),
		)
		fromU = sqlgraph.SetNeighbors(uq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QuerySessions chains the current query on the "sessions" edge.
func (uq *UserQuery) QuerySessions() *SessionQuery {
	query := (&SessionClient{config: uq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := uq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(user.Table, user.FieldID, selector),
			sqlgraph.To(session.Table, session.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, user.SessionsTable, user.SessionsColumn),
		)
		fromU = sqlgraph.SetNeighbors(uq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryProfile chains the current query on the "profile" edge.
func (uq *UserQuery) QueryProfile() *ProfileQuery {
	query := (&ProfileClient{config: uq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := uq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(user.Table, user.FieldID, selector),
			sqlgraph.To(profile.Table, profile.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, user.ProfileTable, user.ProfileColumn),
		)
		fromU = sqlgraph.SetNeighbors(uq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryWorkouts chains the current query on the "workouts" edge.
func (uq *UserQuery) QueryWorkouts() *WorkoutQuery {
	query := (&WorkoutClient{config: uq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := uq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(user.Table, user.FieldID, selector),
			sqlgraph.To(workout.Table, workout.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, user.WorkoutsTable, user.WorkoutsColumn),
		)
		fromU = sqlgraph.SetNeighbors(uq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryWorkoutLogs chains the current query on the "workout_logs" edge.
func (uq *UserQuery) QueryWorkoutLogs() *WorkoutLogQuery {
	query := (&WorkoutLogClient{config: uq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := uq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(user.Table, user.FieldID, selector),
			sqlgraph.To(workoutlog.Table, workoutlog.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, user.WorkoutLogsTable, user.WorkoutLogsColumn),
		)
		fromU = sqlgraph.SetNeighbors(uq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryPrivateToken chains the current query on the "private_token" edge.
func (uq *UserQuery) QueryPrivateToken() *PrivateTokenQuery {
	query := (&PrivateTokenClient{config: uq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := uq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(user.Table, user.FieldID, selector),
			sqlgraph.To(privatetoken.Table, privatetoken.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, user.PrivateTokenTable, user.PrivateTokenColumn),
		)
		fromU = sqlgraph.SetNeighbors(uq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first User entity from the query.
// Returns a *NotFoundError when no User was found.
func (uq *UserQuery) First(ctx context.Context) (*User, error) {
	nodes, err := uq.Limit(1).All(setContextOp(ctx, uq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{user.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (uq *UserQuery) FirstX(ctx context.Context) *User {
	node, err := uq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first User ID from the query.
// Returns a *NotFoundError when no User ID was found.
func (uq *UserQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = uq.Limit(1).IDs(setContextOp(ctx, uq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{user.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (uq *UserQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := uq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single User entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one User entity is found.
// Returns a *NotFoundError when no User entities are found.
func (uq *UserQuery) Only(ctx context.Context) (*User, error) {
	nodes, err := uq.Limit(2).All(setContextOp(ctx, uq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{user.Label}
	default:
		return nil, &NotSingularError{user.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (uq *UserQuery) OnlyX(ctx context.Context) *User {
	node, err := uq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only User ID in the query.
// Returns a *NotSingularError when more than one User ID is found.
// Returns a *NotFoundError when no entities are found.
func (uq *UserQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = uq.Limit(2).IDs(setContextOp(ctx, uq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{user.Label}
	default:
		err = &NotSingularError{user.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (uq *UserQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := uq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Users.
func (uq *UserQuery) All(ctx context.Context) ([]*User, error) {
	ctx = setContextOp(ctx, uq.ctx, ent.OpQueryAll)
	if err := uq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*User, *UserQuery]()
	return withInterceptors[[]*User](ctx, uq, qr, uq.inters)
}

// AllX is like All, but panics if an error occurs.
func (uq *UserQuery) AllX(ctx context.Context) []*User {
	nodes, err := uq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of User IDs.
func (uq *UserQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if uq.ctx.Unique == nil && uq.path != nil {
		uq.Unique(true)
	}
	ctx = setContextOp(ctx, uq.ctx, ent.OpQueryIDs)
	if err = uq.Select(user.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (uq *UserQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := uq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (uq *UserQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, uq.ctx, ent.OpQueryCount)
	if err := uq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, uq, querierCount[*UserQuery](), uq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (uq *UserQuery) CountX(ctx context.Context) int {
	count, err := uq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (uq *UserQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, uq.ctx, ent.OpQueryExist)
	switch _, err := uq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (uq *UserQuery) ExistX(ctx context.Context) bool {
	exist, err := uq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the UserQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (uq *UserQuery) Clone() *UserQuery {
	if uq == nil {
		return nil
	}
	return &UserQuery{
		config:           uq.config,
		ctx:              uq.ctx.Clone(),
		order:            append([]user.OrderOption{}, uq.order...),
		inters:           append([]Interceptor{}, uq.inters...),
		predicates:       append([]predicate.User{}, uq.predicates...),
		withBodyweights:  uq.withBodyweights.Clone(),
		withSessions:     uq.withSessions.Clone(),
		withProfile:      uq.withProfile.Clone(),
		withWorkouts:     uq.withWorkouts.Clone(),
		withWorkoutLogs:  uq.withWorkoutLogs.Clone(),
		withPrivateToken: uq.withPrivateToken.Clone(),
		// clone intermediate query.
		sql:  uq.sql.Clone(),
		path: uq.path,
	}
}

// WithBodyweights tells the query-builder to eager-load the nodes that are connected to
// the "bodyweights" edge. The optional arguments are used to configure the query builder of the edge.
func (uq *UserQuery) WithBodyweights(opts ...func(*BodyweightQuery)) *UserQuery {
	query := (&BodyweightClient{config: uq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	uq.withBodyweights = query
	return uq
}

// WithSessions tells the query-builder to eager-load the nodes that are connected to
// the "sessions" edge. The optional arguments are used to configure the query builder of the edge.
func (uq *UserQuery) WithSessions(opts ...func(*SessionQuery)) *UserQuery {
	query := (&SessionClient{config: uq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	uq.withSessions = query
	return uq
}

// WithProfile tells the query-builder to eager-load the nodes that are connected to
// the "profile" edge. The optional arguments are used to configure the query builder of the edge.
func (uq *UserQuery) WithProfile(opts ...func(*ProfileQuery)) *UserQuery {
	query := (&ProfileClient{config: uq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	uq.withProfile = query
	return uq
}

// WithWorkouts tells the query-builder to eager-load the nodes that are connected to
// the "workouts" edge. The optional arguments are used to configure the query builder of the edge.
func (uq *UserQuery) WithWorkouts(opts ...func(*WorkoutQuery)) *UserQuery {
	query := (&WorkoutClient{config: uq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	uq.withWorkouts = query
	return uq
}

// WithWorkoutLogs tells the query-builder to eager-load the nodes that are connected to
// the "workout_logs" edge. The optional arguments are used to configure the query builder of the edge.
func (uq *UserQuery) WithWorkoutLogs(opts ...func(*WorkoutLogQuery)) *UserQuery {
	query := (&WorkoutLogClient{config: uq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	uq.withWorkoutLogs = query
	return uq
}

// WithPrivateToken tells the query-builder to eager-load the nodes that are connected to
// the "private_token" edge. The optional arguments are used to configure the query builder of the edge.
func (uq *UserQuery) WithPrivateToken(opts ...func(*PrivateTokenQuery)) *UserQuery {
	query := (&PrivateTokenClient{config: uq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	uq.withPrivateToken = query
	return uq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		CreatedAt time.Time `json:"created_at,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.User.Query().
//		GroupBy(user.FieldCreatedAt).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (uq *UserQuery) GroupBy(field string, fields ...string) *UserGroupBy {
	uq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &UserGroupBy{build: uq}
	grbuild.flds = &uq.ctx.Fields
	grbuild.label = user.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		CreatedAt time.Time `json:"created_at,omitempty"`
//	}
//
//	client.User.Query().
//		Select(user.FieldCreatedAt).
//		Scan(ctx, &v)
func (uq *UserQuery) Select(fields ...string) *UserSelect {
	uq.ctx.Fields = append(uq.ctx.Fields, fields...)
	sbuild := &UserSelect{UserQuery: uq}
	sbuild.label = user.Label
	sbuild.flds, sbuild.scan = &uq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a UserSelect configured with the given aggregations.
func (uq *UserQuery) Aggregate(fns ...AggregateFunc) *UserSelect {
	return uq.Select().Aggregate(fns...)
}

func (uq *UserQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range uq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, uq); err != nil {
				return err
			}
		}
	}
	for _, f := range uq.ctx.Fields {
		if !user.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if uq.path != nil {
		prev, err := uq.path(ctx)
		if err != nil {
			return err
		}
		uq.sql = prev
	}
	return nil
}

func (uq *UserQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*User, error) {
	var (
		nodes       = []*User{}
		_spec       = uq.querySpec()
		loadedTypes = [6]bool{
			uq.withBodyweights != nil,
			uq.withSessions != nil,
			uq.withProfile != nil,
			uq.withWorkouts != nil,
			uq.withWorkoutLogs != nil,
			uq.withPrivateToken != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*User).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &User{config: uq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, uq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := uq.withBodyweights; query != nil {
		if err := uq.loadBodyweights(ctx, query, nodes,
			func(n *User) { n.Edges.Bodyweights = []*Bodyweight{} },
			func(n *User, e *Bodyweight) { n.Edges.Bodyweights = append(n.Edges.Bodyweights, e) }); err != nil {
			return nil, err
		}
	}
	if query := uq.withSessions; query != nil {
		if err := uq.loadSessions(ctx, query, nodes,
			func(n *User) { n.Edges.Sessions = []*Session{} },
			func(n *User, e *Session) { n.Edges.Sessions = append(n.Edges.Sessions, e) }); err != nil {
			return nil, err
		}
	}
	if query := uq.withProfile; query != nil {
		if err := uq.loadProfile(ctx, query, nodes, nil,
			func(n *User, e *Profile) { n.Edges.Profile = e }); err != nil {
			return nil, err
		}
	}
	if query := uq.withWorkouts; query != nil {
		if err := uq.loadWorkouts(ctx, query, nodes,
			func(n *User) { n.Edges.Workouts = []*Workout{} },
			func(n *User, e *Workout) { n.Edges.Workouts = append(n.Edges.Workouts, e) }); err != nil {
			return nil, err
		}
	}
	if query := uq.withWorkoutLogs; query != nil {
		if err := uq.loadWorkoutLogs(ctx, query, nodes,
			func(n *User) { n.Edges.WorkoutLogs = []*WorkoutLog{} },
			func(n *User, e *WorkoutLog) { n.Edges.WorkoutLogs = append(n.Edges.WorkoutLogs, e) }); err != nil {
			return nil, err
		}
	}
	if query := uq.withPrivateToken; query != nil {
		if err := uq.loadPrivateToken(ctx, query, nodes,
			func(n *User) { n.Edges.PrivateToken = []*PrivateToken{} },
			func(n *User, e *PrivateToken) { n.Edges.PrivateToken = append(n.Edges.PrivateToken, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (uq *UserQuery) loadBodyweights(ctx context.Context, query *BodyweightQuery, nodes []*User, init func(*User), assign func(*User, *Bodyweight)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*User)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Bodyweight(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(user.BodyweightsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.user_bodyweights
		if fk == nil {
			return fmt.Errorf(`foreign-key "user_bodyweights" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "user_bodyweights" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (uq *UserQuery) loadSessions(ctx context.Context, query *SessionQuery, nodes []*User, init func(*User), assign func(*User, *Session)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*User)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Session(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(user.SessionsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.user_sessions
		if fk == nil {
			return fmt.Errorf(`foreign-key "user_sessions" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "user_sessions" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (uq *UserQuery) loadProfile(ctx context.Context, query *ProfileQuery, nodes []*User, init func(*User), assign func(*User, *Profile)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*User)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
	}
	query.withFKs = true
	query.Where(predicate.Profile(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(user.ProfileColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.user_profile
		if fk == nil {
			return fmt.Errorf(`foreign-key "user_profile" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "user_profile" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (uq *UserQuery) loadWorkouts(ctx context.Context, query *WorkoutQuery, nodes []*User, init func(*User), assign func(*User, *Workout)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*User)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Workout(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(user.WorkoutsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.user_workouts
		if fk == nil {
			return fmt.Errorf(`foreign-key "user_workouts" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "user_workouts" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (uq *UserQuery) loadWorkoutLogs(ctx context.Context, query *WorkoutLogQuery, nodes []*User, init func(*User), assign func(*User, *WorkoutLog)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*User)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.WorkoutLog(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(user.WorkoutLogsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.user_workout_logs
		if fk == nil {
			return fmt.Errorf(`foreign-key "user_workout_logs" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "user_workout_logs" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (uq *UserQuery) loadPrivateToken(ctx context.Context, query *PrivateTokenQuery, nodes []*User, init func(*User), assign func(*User, *PrivateToken)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*User)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.PrivateToken(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(user.PrivateTokenColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.user_private_token
		if fk == nil {
			return fmt.Errorf(`foreign-key "user_private_token" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "user_private_token" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (uq *UserQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := uq.querySpec()
	_spec.Node.Columns = uq.ctx.Fields
	if len(uq.ctx.Fields) > 0 {
		_spec.Unique = uq.ctx.Unique != nil && *uq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, uq.driver, _spec)
}

func (uq *UserQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(user.Table, user.Columns, sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID))
	_spec.From = uq.sql
	if unique := uq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if uq.path != nil {
		_spec.Unique = true
	}
	if fields := uq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, user.FieldID)
		for i := range fields {
			if fields[i] != user.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := uq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := uq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := uq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := uq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (uq *UserQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(uq.driver.Dialect())
	t1 := builder.Table(user.Table)
	columns := uq.ctx.Fields
	if len(columns) == 0 {
		columns = user.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if uq.sql != nil {
		selector = uq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if uq.ctx.Unique != nil && *uq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range uq.predicates {
		p(selector)
	}
	for _, p := range uq.order {
		p(selector)
	}
	if offset := uq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := uq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// UserGroupBy is the group-by builder for User entities.
type UserGroupBy struct {
	selector
	build *UserQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (ugb *UserGroupBy) Aggregate(fns ...AggregateFunc) *UserGroupBy {
	ugb.fns = append(ugb.fns, fns...)
	return ugb
}

// Scan applies the selector query and scans the result into the given value.
func (ugb *UserGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ugb.build.ctx, ent.OpQueryGroupBy)
	if err := ugb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*UserQuery, *UserGroupBy](ctx, ugb.build, ugb, ugb.build.inters, v)
}

func (ugb *UserGroupBy) sqlScan(ctx context.Context, root *UserQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(ugb.fns))
	for _, fn := range ugb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*ugb.flds)+len(ugb.fns))
		for _, f := range *ugb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*ugb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ugb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// UserSelect is the builder for selecting fields of User entities.
type UserSelect struct {
	*UserQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (us *UserSelect) Aggregate(fns ...AggregateFunc) *UserSelect {
	us.fns = append(us.fns, fns...)
	return us
}

// Scan applies the selector query and scans the result into the given value.
func (us *UserSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, us.ctx, ent.OpQuerySelect)
	if err := us.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*UserQuery, *UserSelect](ctx, us.UserQuery, us, us.inters, v)
}

func (us *UserSelect) sqlScan(ctx context.Context, root *UserQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(us.fns))
	for _, fn := range us.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*us.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := us.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
