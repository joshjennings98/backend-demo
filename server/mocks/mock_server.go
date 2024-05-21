// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/joshjennings98/backend-demo/server/server (interfaces: IPresentation,IPresentationServer,ICommandManager)
//
// Generated by this command:
//
//	mockgen -destination=../mocks/mock_server.go -package=mocks github.com/joshjennings98/backend-demo/server/server IPresentation,IPresentationServer,ICommandManager
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	http "net/http"
	reflect "reflect"

	websocket "github.com/gorilla/websocket"
	types "github.com/joshjennings98/backend-demo/server/types"
	gomock "go.uber.org/mock/gomock"
)

// MockIPresentation is a mock of IPresentation interface.
type MockIPresentation struct {
	ctrl     *gomock.Controller
	recorder *MockIPresentationMockRecorder
}

// MockIPresentationMockRecorder is the mock recorder for MockIPresentation.
type MockIPresentationMockRecorder struct {
	mock *MockIPresentation
}

// NewMockIPresentation creates a new mock instance.
func NewMockIPresentation(ctrl *gomock.Controller) *MockIPresentation {
	mock := &MockIPresentation{ctrl: ctrl}
	mock.recorder = &MockIPresentationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIPresentation) EXPECT() *MockIPresentationMockRecorder {
	return m.recorder
}

// GetPreCommands mocks base method.
func (m *MockIPresentation) GetPreCommands() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPreCommands")
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetPreCommands indicates an expected call of GetPreCommands.
func (mr *MockIPresentationMockRecorder) GetPreCommands() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPreCommands", reflect.TypeOf((*MockIPresentation)(nil).GetPreCommands))
}

// GetSlide mocks base method.
func (m *MockIPresentation) GetSlide(arg0 int) types.Slide {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSlide", arg0)
	ret0, _ := ret[0].(types.Slide)
	return ret0
}

// GetSlide indicates an expected call of GetSlide.
func (mr *MockIPresentationMockRecorder) GetSlide(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSlide", reflect.TypeOf((*MockIPresentation)(nil).GetSlide), arg0)
}

// GetSlideCount mocks base method.
func (m *MockIPresentation) GetSlideCount() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSlideCount")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetSlideCount indicates an expected call of GetSlideCount.
func (mr *MockIPresentationMockRecorder) GetSlideCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSlideCount", reflect.TypeOf((*MockIPresentation)(nil).GetSlideCount))
}

// Initialise mocks base method.
func (m *MockIPresentation) Initialise(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Initialise", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Initialise indicates an expected call of Initialise.
func (mr *MockIPresentationMockRecorder) Initialise(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initialise", reflect.TypeOf((*MockIPresentation)(nil).Initialise), arg0)
}

// ParsePreCommands mocks base method.
func (m *MockIPresentation) ParsePreCommands(arg0 []string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParsePreCommands", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParsePreCommands indicates an expected call of ParsePreCommands.
func (mr *MockIPresentationMockRecorder) ParsePreCommands(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParsePreCommands", reflect.TypeOf((*MockIPresentation)(nil).ParsePreCommands), arg0)
}

// ParseSlide mocks base method.
func (m *MockIPresentation) ParseSlide(arg0 string, arg1 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ParseSlide", arg0, arg1)
}

// ParseSlide indicates an expected call of ParseSlide.
func (mr *MockIPresentationMockRecorder) ParseSlide(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseSlide", reflect.TypeOf((*MockIPresentation)(nil).ParseSlide), arg0, arg1)
}

// ParseSlides mocks base method.
func (m *MockIPresentation) ParseSlides(arg0 []string, arg1 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseSlides", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ParseSlides indicates an expected call of ParseSlides.
func (mr *MockIPresentationMockRecorder) ParseSlides(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseSlides", reflect.TypeOf((*MockIPresentation)(nil).ParseSlides), arg0, arg1)
}

// MockIPresentationServer is a mock of IPresentationServer interface.
type MockIPresentationServer struct {
	ctrl     *gomock.Controller
	recorder *MockIPresentationServerMockRecorder
}

// MockIPresentationServerMockRecorder is the mock recorder for MockIPresentationServer.
type MockIPresentationServerMockRecorder struct {
	mock *MockIPresentationServer
}

// NewMockIPresentationServer creates a new mock instance.
func NewMockIPresentationServer(ctrl *gomock.Controller) *MockIPresentationServer {
	mock := &MockIPresentationServer{ctrl: ctrl}
	mock.recorder = &MockIPresentationServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIPresentationServer) EXPECT() *MockIPresentationServerMockRecorder {
	return m.recorder
}

// GetPreCommands mocks base method.
func (m *MockIPresentationServer) GetPreCommands() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPreCommands")
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetPreCommands indicates an expected call of GetPreCommands.
func (mr *MockIPresentationServerMockRecorder) GetPreCommands() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPreCommands", reflect.TypeOf((*MockIPresentationServer)(nil).GetPreCommands))
}

// GetSlide mocks base method.
func (m *MockIPresentationServer) GetSlide(arg0 int) types.Slide {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSlide", arg0)
	ret0, _ := ret[0].(types.Slide)
	return ret0
}

// GetSlide indicates an expected call of GetSlide.
func (mr *MockIPresentationServerMockRecorder) GetSlide(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSlide", reflect.TypeOf((*MockIPresentationServer)(nil).GetSlide), arg0)
}

// GetSlideCount mocks base method.
func (m *MockIPresentationServer) GetSlideCount() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSlideCount")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetSlideCount indicates an expected call of GetSlideCount.
func (mr *MockIPresentationServerMockRecorder) GetSlideCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSlideCount", reflect.TypeOf((*MockIPresentationServer)(nil).GetSlideCount))
}

// HandlerCommandStart mocks base method.
func (m *MockIPresentationServer) HandlerCommandStart(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlerCommandStart", arg0, arg1)
}

// HandlerCommandStart indicates an expected call of HandlerCommandStart.
func (mr *MockIPresentationServerMockRecorder) HandlerCommandStart(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlerCommandStart", reflect.TypeOf((*MockIPresentationServer)(nil).HandlerCommandStart), arg0, arg1)
}

// HandlerCommandStatus mocks base method.
func (m *MockIPresentationServer) HandlerCommandStatus(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlerCommandStatus", arg0, arg1)
}

// HandlerCommandStatus indicates an expected call of HandlerCommandStatus.
func (mr *MockIPresentationServerMockRecorder) HandlerCommandStatus(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlerCommandStatus", reflect.TypeOf((*MockIPresentationServer)(nil).HandlerCommandStatus), arg0, arg1)
}

// HandlerCommandStop mocks base method.
func (m *MockIPresentationServer) HandlerCommandStop(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlerCommandStop", arg0, arg1)
}

// HandlerCommandStop indicates an expected call of HandlerCommandStop.
func (mr *MockIPresentationServerMockRecorder) HandlerCommandStop(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlerCommandStop", reflect.TypeOf((*MockIPresentationServer)(nil).HandlerCommandStop), arg0, arg1)
}

// HandlerIndex mocks base method.
func (m *MockIPresentationServer) HandlerIndex(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlerIndex", arg0, arg1)
}

// HandlerIndex indicates an expected call of HandlerIndex.
func (mr *MockIPresentationServerMockRecorder) HandlerIndex(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlerIndex", reflect.TypeOf((*MockIPresentationServer)(nil).HandlerIndex), arg0, arg1)
}

// HandlerInit mocks base method.
func (m *MockIPresentationServer) HandlerInit(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlerInit", arg0, arg1)
}

// HandlerInit indicates an expected call of HandlerInit.
func (mr *MockIPresentationServerMockRecorder) HandlerInit(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlerInit", reflect.TypeOf((*MockIPresentationServer)(nil).HandlerInit), arg0, arg1)
}

// HandlerSlideByIndex mocks base method.
func (m *MockIPresentationServer) HandlerSlideByIndex(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlerSlideByIndex", arg0, arg1)
}

// HandlerSlideByIndex indicates an expected call of HandlerSlideByIndex.
func (mr *MockIPresentationServerMockRecorder) HandlerSlideByIndex(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlerSlideByIndex", reflect.TypeOf((*MockIPresentationServer)(nil).HandlerSlideByIndex), arg0, arg1)
}

// HandlerSlideByQuery mocks base method.
func (m *MockIPresentationServer) HandlerSlideByQuery(arg0 http.ResponseWriter, arg1 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "HandlerSlideByQuery", arg0, arg1)
}

// HandlerSlideByQuery indicates an expected call of HandlerSlideByQuery.
func (mr *MockIPresentationServerMockRecorder) HandlerSlideByQuery(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandlerSlideByQuery", reflect.TypeOf((*MockIPresentationServer)(nil).HandlerSlideByQuery), arg0, arg1)
}

// Initialise mocks base method.
func (m *MockIPresentationServer) Initialise(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Initialise", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Initialise indicates an expected call of Initialise.
func (mr *MockIPresentationServerMockRecorder) Initialise(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initialise", reflect.TypeOf((*MockIPresentationServer)(nil).Initialise), arg0)
}

// ParsePreCommands mocks base method.
func (m *MockIPresentationServer) ParsePreCommands(arg0 []string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParsePreCommands", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParsePreCommands indicates an expected call of ParsePreCommands.
func (mr *MockIPresentationServerMockRecorder) ParsePreCommands(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParsePreCommands", reflect.TypeOf((*MockIPresentationServer)(nil).ParsePreCommands), arg0)
}

// ParseSlide mocks base method.
func (m *MockIPresentationServer) ParseSlide(arg0 string, arg1 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ParseSlide", arg0, arg1)
}

// ParseSlide indicates an expected call of ParseSlide.
func (mr *MockIPresentationServerMockRecorder) ParseSlide(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseSlide", reflect.TypeOf((*MockIPresentationServer)(nil).ParseSlide), arg0, arg1)
}

// ParseSlides mocks base method.
func (m *MockIPresentationServer) ParseSlides(arg0 []string, arg1 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseSlides", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ParseSlides indicates an expected call of ParseSlides.
func (mr *MockIPresentationServerMockRecorder) ParseSlides(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseSlides", reflect.TypeOf((*MockIPresentationServer)(nil).ParseSlides), arg0, arg1)
}

// Start mocks base method.
func (m *MockIPresentationServer) Start(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockIPresentationServerMockRecorder) Start(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockIPresentationServer)(nil).Start), arg0)
}

// MockICommandManager is a mock of ICommandManager interface.
type MockICommandManager struct {
	ctrl     *gomock.Controller
	recorder *MockICommandManagerMockRecorder
}

// MockICommandManagerMockRecorder is the mock recorder for MockICommandManager.
type MockICommandManagerMockRecorder struct {
	mock *MockICommandManager
}

// NewMockICommandManager creates a new mock instance.
func NewMockICommandManager(ctrl *gomock.Controller) *MockICommandManager {
	mock := &MockICommandManager{ctrl: ctrl}
	mock.recorder = &MockICommandManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockICommandManager) EXPECT() *MockICommandManagerMockRecorder {
	return m.recorder
}

// ExecuteCommand mocks base method.
func (m *MockICommandManager) ExecuteCommand(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecuteCommand", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ExecuteCommand indicates an expected call of ExecuteCommand.
func (mr *MockICommandManagerMockRecorder) ExecuteCommand(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecuteCommand", reflect.TypeOf((*MockICommandManager)(nil).ExecuteCommand), arg0, arg1)
}

// GetWebsocketConnection mocks base method.
func (m *MockICommandManager) GetWebsocketConnection() *websocket.Conn {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWebsocketConnection")
	ret0, _ := ret[0].(*websocket.Conn)
	return ret0
}

// GetWebsocketConnection indicates an expected call of GetWebsocketConnection.
func (mr *MockICommandManagerMockRecorder) GetWebsocketConnection() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWebsocketConnection", reflect.TypeOf((*MockICommandManager)(nil).GetWebsocketConnection))
}

// IsRunning mocks base method.
func (m *MockICommandManager) IsRunning() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsRunning")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsRunning indicates an expected call of IsRunning.
func (mr *MockICommandManagerMockRecorder) IsRunning() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsRunning", reflect.TypeOf((*MockICommandManager)(nil).IsRunning))
}

// SetCancelCommand mocks base method.
func (m *MockICommandManager) SetCancelCommand(arg0 context.CancelFunc) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetCancelCommand", arg0)
}

// SetCancelCommand indicates an expected call of SetCancelCommand.
func (mr *MockICommandManagerMockRecorder) SetCancelCommand(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCancelCommand", reflect.TypeOf((*MockICommandManager)(nil).SetCancelCommand), arg0)
}

// SetRunning mocks base method.
func (m *MockICommandManager) SetRunning(arg0 bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetRunning", arg0)
}

// SetRunning indicates an expected call of SetRunning.
func (mr *MockICommandManagerMockRecorder) SetRunning(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRunning", reflect.TypeOf((*MockICommandManager)(nil).SetRunning), arg0)
}

// SetWebsocketConnection mocks base method.
func (m *MockICommandManager) SetWebsocketConnection(arg0 *websocket.Conn) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetWebsocketConnection", arg0)
}

// SetWebsocketConnection indicates an expected call of SetWebsocketConnection.
func (mr *MockICommandManagerMockRecorder) SetWebsocketConnection(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWebsocketConnection", reflect.TypeOf((*MockICommandManager)(nil).SetWebsocketConnection), arg0)
}

// StartCommand mocks base method.
func (m *MockICommandManager) StartCommand(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartCommand", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// StartCommand indicates an expected call of StartCommand.
func (mr *MockICommandManagerMockRecorder) StartCommand(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartCommand", reflect.TypeOf((*MockICommandManager)(nil).StartCommand), arg0)
}

// StopCurrentCommand mocks base method.
func (m *MockICommandManager) StopCurrentCommand() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopCurrentCommand")
	ret0, _ := ret[0].(error)
	return ret0
}

// StopCurrentCommand indicates an expected call of StopCurrentCommand.
func (mr *MockICommandManagerMockRecorder) StopCurrentCommand() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopCurrentCommand", reflect.TypeOf((*MockICommandManager)(nil).StopCurrentCommand))
}

// TermClear mocks base method.
func (m *MockICommandManager) TermClear() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TermClear")
	ret0, _ := ret[0].(error)
	return ret0
}

// TermClear indicates an expected call of TermClear.
func (mr *MockICommandManagerMockRecorder) TermClear() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TermClear", reflect.TypeOf((*MockICommandManager)(nil).TermClear))
}

// TermMessage mocks base method.
func (m *MockICommandManager) TermMessage(arg0 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TermMessage", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// TermMessage indicates an expected call of TermMessage.
func (mr *MockICommandManagerMockRecorder) TermMessage(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TermMessage", reflect.TypeOf((*MockICommandManager)(nil).TermMessage), arg0)
}