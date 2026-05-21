package users

import "github.com/labstack/echo/v4"

// Handler defines user HTTP handlers.
type Handler interface {
	Create(c echo.Context) error
	Get(c echo.Context) error
	List(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error
}

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

func (h *handler) Create(c echo.Context) error {
	return nil
}

func (h *handler) Get(c echo.Context) error {
	return nil
}

func (h *handler) List(c echo.Context) error {
	return nil
}

func (h *handler) Update(c echo.Context) error {
	return nil
}

func (h *handler) Delete(c echo.Context) error {
	return nil
}
