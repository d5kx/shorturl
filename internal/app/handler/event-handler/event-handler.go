package eventhandler

type Handler struct {
}

func (h *Handler) Run() error {
	return nil
}

func New() Handler {
	return Handler{}
}
