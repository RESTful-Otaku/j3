package api

// EnableTestBypass is only compiled into the test binary. It allows in-process
// API tests to skip bearer-token lookup without leaving a production backdoor.
func (s *Server) EnableTestBypass() {
	s.testToken = "test-only"
}
