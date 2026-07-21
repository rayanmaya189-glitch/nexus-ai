package testutil

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aeroxe/nexus-backend/internal/domain/entities"
)

type MockUserRepository struct {
	mu     sync.Mutex
	users  map[int64]*entities.User
	nextID int64
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{users: make(map[int64]*entities.User), nextID: 1}
}

func (m *MockUserRepository) FindByID(_ context.Context, id int64) (*entities.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockUserRepository) FindByEmail(_ context.Context, email string) (*entities.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockUserRepository) FindByTenantID(_ context.Context, tenantID int64) ([]*entities.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.User
	for _, u := range m.users {
		if u.TenantID == tenantID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *MockUserRepository) Create(_ context.Context, user *entities.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	user.ID = m.nextID
	m.nextID++
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Update(_ context.Context, user *entities.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) Count(_ context.Context, tenantID int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for _, u := range m.users {
		if u.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

type MockTenantRepository struct {
	mu      sync.Mutex
	tenants map[int64]*entities.Tenant
	nextID  int64
}

func NewMockTenantRepository() *MockTenantRepository {
	return &MockTenantRepository{tenants: make(map[int64]*entities.Tenant), nextID: 1}
}

func (m *MockTenantRepository) FindByID(_ context.Context, id int64) (*entities.Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.tenants[id]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockTenantRepository) FindBySlug(_ context.Context, slug string) (*entities.Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tenants {
		if t.Slug == slug {
			return t, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockTenantRepository) Create(_ context.Context, tenant *entities.Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	tenant.ID = m.nextID
	m.nextID++
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = time.Now()
	}
	tenant.UpdatedAt = time.Now()
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *MockTenantRepository) Update(_ context.Context, tenant *entities.Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tenants[tenant.ID] = tenant
	return nil
}

func (m *MockTenantRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tenants, id)
	return nil
}

func (m *MockTenantRepository) List(_ context.Context, page, perPage int) ([]*entities.Tenant, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []*entities.Tenant
	for _, t := range m.tenants {
		all = append(all, t)
	}
	total := int64(len(all))
	start := (page - 1) * perPage
	if start >= len(all) {
		return nil, total, nil
	}
	end := start + perPage
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

type MockRoleRepository struct {
	mu      sync.Mutex
	roles   map[int64]*entities.Role
	nextID  int64
	assignments map[string]bool
}

func NewMockRoleRepository() *MockRoleRepository {
	return &MockRoleRepository{roles: make(map[int64]*entities.Role), assignments: make(map[string]bool), nextID: 1}
}

func (m *MockRoleRepository) FindByID(_ context.Context, id int64) (*entities.Role, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.roles[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockRoleRepository) FindByTenantID(_ context.Context, tenantID int64) ([]*entities.Role, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.Role
	for _, r := range m.roles {
		if r.TenantID == tenantID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *MockRoleRepository) Create(_ context.Context, role *entities.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	role.ID = m.nextID
	m.nextID++
	if role.CreatedAt.IsZero() {
		role.CreatedAt = time.Now()
	}
	m.roles[role.ID] = role
	return nil
}

func (m *MockRoleRepository) Update(_ context.Context, role *entities.Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roles[role.ID] = role
	return nil
}

func (m *MockRoleRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.roles, id)
	return nil
}

func (m *MockRoleRepository) AssignRole(_ context.Context, userID, roleID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%d", userID, roleID)
	m.assignments[key] = true
	return nil
}

func (m *MockRoleRepository) RemoveRole(_ context.Context, userID, roleID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%d:%d", userID, roleID)
	delete(m.assignments, key)
	return nil
}

type MockAgentRepository struct {
	mu      sync.Mutex
	agents  map[int64]*entities.Agent
	nextID  int64
}

func NewMockAgentRepository() *MockAgentRepository {
	return &MockAgentRepository{agents: make(map[int64]*entities.Agent), nextID: 1}
}

func (m *MockAgentRepository) FindByID(_ context.Context, id int64) (*entities.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.agents[id]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockAgentRepository) FindByTenantID(_ context.Context, tenantID int64) ([]*entities.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.Agent
	for _, a := range m.agents {
		if a.TenantID == tenantID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockAgentRepository) FindByType(_ context.Context, tenantID int64, agentType string) ([]*entities.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.Agent
	for _, a := range m.agents {
		if a.TenantID == tenantID && a.AgentType == agentType {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *MockAgentRepository) Create(_ context.Context, agent *entities.Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	agent.ID = m.nextID
	m.nextID++
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = time.Now()
	}
	agent.UpdatedAt = time.Now()
	m.agents[agent.ID] = agent
	return nil
}

func (m *MockAgentRepository) Update(_ context.Context, agent *entities.Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agent.ID] = agent
	return nil
}

func (m *MockAgentRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.agents, id)
	return nil
}

type MockAgentExecutionRepository struct {
	mu         sync.Mutex
	executions map[int64]*entities.AgentExecution
	nextID     int64
}

func NewMockAgentExecutionRepository() *MockAgentExecutionRepository {
	return &MockAgentExecutionRepository{executions: make(map[int64]*entities.AgentExecution), nextID: 1}
}

func (m *MockAgentExecutionRepository) FindByID(_ context.Context, id int64) (*entities.AgentExecution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.executions[id]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockAgentExecutionRepository) FindByAgentID(_ context.Context, agentID int64) ([]*entities.AgentExecution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.AgentExecution
	for _, e := range m.executions {
		if e.AgentID == agentID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *MockAgentExecutionRepository) Create(_ context.Context, execution *entities.AgentExecution) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	execution.ID = m.nextID
	m.nextID++
	m.executions[execution.ID] = execution
	return nil
}

func (m *MockAgentExecutionRepository) Update(_ context.Context, execution *entities.AgentExecution) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions[execution.ID] = execution
	return nil
}

type MockAgentStepRepository struct {
	mu      sync.Mutex
	steps   map[int64]*entities.AgentStep
	nextID  int64
}

func NewMockAgentStepRepository() *MockAgentStepRepository {
	return &MockAgentStepRepository{steps: make(map[int64]*entities.AgentStep), nextID: 1}
}

func (m *MockAgentStepRepository) FindByExecutionID(_ context.Context, executionID int64) ([]*entities.AgentStep, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.AgentStep
	for _, s := range m.steps {
		if s.ExecutionID == executionID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockAgentStepRepository) Create(_ context.Context, step *entities.AgentStep) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	step.ID = m.nextID
	m.nextID++
	m.steps[step.ID] = step
	return nil
}

func (m *MockAgentStepRepository) Update(_ context.Context, step *entities.AgentStep) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.steps[step.ID] = step
	return nil
}

type MockDocumentRepository struct {
	mu      sync.Mutex
	docs    map[int64]*entities.Document
	nextID  int64
}

func NewMockDocumentRepository() *MockDocumentRepository {
	return &MockDocumentRepository{docs: make(map[int64]*entities.Document), nextID: 1}
}

func (m *MockDocumentRepository) FindByID(_ context.Context, id int64) (*entities.Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.docs[id]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockDocumentRepository) FindByTenantID(_ context.Context, tenantID int64) ([]*entities.Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.Document
	for _, d := range m.docs {
		if d.TenantID == tenantID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (m *MockDocumentRepository) Create(_ context.Context, doc *entities.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	doc.ID = m.nextID
	m.nextID++
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}
	doc.UpdatedAt = time.Now()
	m.docs[doc.ID] = doc
	return nil
}

func (m *MockDocumentRepository) Update(_ context.Context, doc *entities.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.docs[doc.ID] = doc
	return nil
}

func (m *MockDocumentRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.docs, id)
	return nil
}

func (m *MockDocumentRepository) Count(_ context.Context, tenantID int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for _, d := range m.docs {
		if d.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}

type MockDocumentChunkRepository struct {
	mu      sync.Mutex
	chunks  map[int64]*entities.DocumentChunk
	nextID  int64
}

func NewMockDocumentChunkRepository() *MockDocumentChunkRepository {
	return &MockDocumentChunkRepository{chunks: make(map[int64]*entities.DocumentChunk), nextID: 1}
}

func (m *MockDocumentChunkRepository) FindByID(_ context.Context, id int64) (*entities.DocumentChunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.chunks[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockDocumentChunkRepository) FindByDocumentID(_ context.Context, documentID int64) ([]*entities.DocumentChunk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.DocumentChunk
	for _, c := range m.chunks {
		if c.DocumentID == documentID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *MockDocumentChunkRepository) Create(_ context.Context, chunk *entities.DocumentChunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	chunk.ID = m.nextID
	m.nextID++
	if chunk.CreatedAt.IsZero() {
		chunk.CreatedAt = time.Now()
	}
	m.chunks[chunk.ID] = chunk
	return nil
}

func (m *MockDocumentChunkRepository) CreateBatch(_ context.Context, chunks []*entities.DocumentChunk) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, chunk := range chunks {
		chunk.ID = m.nextID
		m.nextID++
		if chunk.CreatedAt.IsZero() {
			chunk.CreatedAt = time.Now()
		}
		m.chunks[chunk.ID] = chunk
	}
	return nil
}

func (m *MockDocumentChunkRepository) DeleteByDocumentID(_ context.Context, documentID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, c := range m.chunks {
		if c.DocumentID == documentID {
			delete(m.chunks, id)
		}
	}
	return nil
}

type MockDocumentSetRepository struct {
	mu      sync.Mutex
	sets    map[int64]*entities.DocumentSet
	nextID  int64
}

func NewMockDocumentSetRepository() *MockDocumentSetRepository {
	return &MockDocumentSetRepository{sets: make(map[int64]*entities.DocumentSet), nextID: 1}
}

func (m *MockDocumentSetRepository) FindByID(_ context.Context, id int64) (*entities.DocumentSet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sets[id]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockDocumentSetRepository) FindByTenantID(_ context.Context, tenantID int64) ([]*entities.DocumentSet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.DocumentSet
	for _, s := range m.sets {
		if s.TenantID == tenantID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockDocumentSetRepository) Create(_ context.Context, set *entities.DocumentSet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	set.ID = m.nextID
	m.nextID++
	if set.CreatedAt.IsZero() {
		set.CreatedAt = time.Now()
	}
	set.UpdatedAt = time.Now()
	m.sets[set.ID] = set
	return nil
}

func (m *MockDocumentSetRepository) Update(_ context.Context, set *entities.DocumentSet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sets[set.ID] = set
	return nil
}

func (m *MockDocumentSetRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sets, id)
	return nil
}

type MockWorkflowRepository struct {
	mu        sync.Mutex
	workflows map[int64]*entities.Workflow
	nextID    int64
}

func NewMockWorkflowRepository() *MockWorkflowRepository {
	return &MockWorkflowRepository{workflows: make(map[int64]*entities.Workflow), nextID: 1}
}

func (m *MockWorkflowRepository) FindByID(_ context.Context, id int64) (*entities.Workflow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w, ok := m.workflows[id]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockWorkflowRepository) FindByTenantID(_ context.Context, tenantID int64) ([]*entities.Workflow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.Workflow
	for _, w := range m.workflows {
		if w.TenantID == tenantID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *MockWorkflowRepository) Create(_ context.Context, wf *entities.Workflow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	wf.ID = m.nextID
	m.nextID++
	if wf.CreatedAt.IsZero() {
		wf.CreatedAt = time.Now()
	}
	wf.UpdatedAt = time.Now()
	m.workflows[wf.ID] = wf
	return nil
}

func (m *MockWorkflowRepository) Update(_ context.Context, wf *entities.Workflow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflows[wf.ID] = wf
	return nil
}

func (m *MockWorkflowRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.workflows, id)
	return nil
}

type MockWorkflowStepRepository struct {
	mu      sync.Mutex
	steps   map[int64]*entities.WorkflowStep
	nextID  int64
}

func NewMockWorkflowStepRepository() *MockWorkflowStepRepository {
	return &MockWorkflowStepRepository{steps: make(map[int64]*entities.WorkflowStep), nextID: 1}
}

func (m *MockWorkflowStepRepository) FindByWorkflowID(_ context.Context, workflowID int64) ([]*entities.WorkflowStep, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*entities.WorkflowStep
	for _, s := range m.steps {
		if s.WorkflowID == workflowID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockWorkflowStepRepository) Create(_ context.Context, step *entities.WorkflowStep) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	step.ID = m.nextID
	m.nextID++
	if step.CreatedAt.IsZero() {
		step.CreatedAt = time.Now()
	}
	m.steps[step.ID] = step
	return nil
}

func (m *MockWorkflowStepRepository) Update(_ context.Context, step *entities.WorkflowStep) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.steps[step.ID] = step
	return nil
}

type MockAuditLogRepository struct {
	mu      sync.Mutex
	logs    map[int64]*entities.AuditLog
	nextID  int64
}

func NewMockAuditLogRepository() *MockAuditLogRepository {
	return &MockAuditLogRepository{logs: make(map[int64]*entities.AuditLog), nextID: 1}
}

func (m *MockAuditLogRepository) FindByID(_ context.Context, id int64) (*entities.AuditLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l, ok := m.logs[id]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockAuditLogRepository) FindByTenantID(_ context.Context, tenantID int64, limit, offset int) ([]*entities.AuditLog, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var all []*entities.AuditLog
	for _, l := range m.logs {
		if l.TenantID == tenantID {
			all = append(all, l)
		}
	}
	total := int64(len(all))
	start := offset
	if start >= len(all) {
		return nil, total, nil
	}
	end := start + limit
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func (m *MockAuditLogRepository) Create(_ context.Context, log *entities.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.ID = m.nextID
	m.nextID++
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	m.logs[log.ID] = log
	return nil
}

func (m *MockAuditLogRepository) Count(_ context.Context, tenantID int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	for _, l := range m.logs {
		if l.TenantID == tenantID {
			count++
		}
	}
	return count, nil
}
