package serverpool

func (s *ServerPool) SetMaintenance(
	url string,
	maintenance bool,
) {

	s.mux.Lock()
	defer s.mux.Unlock()

	for _, backend := range s.backends {

		if backend.URL.String() == url {

			backend.SetMaintenance(maintenance)

			return
		}
	}
}
