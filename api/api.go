package api

import (
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/luizbafilho/fusis/fusis"
)

// ApiService ...
type ApiService struct {
	echo     *echo.Echo
	balancer fusis.Balancer
	env      string
}

//NewAPI ...
func NewAPI(balancer fusis.Balancer) ApiService {
	as := ApiService{
		echo:     echo.New(),
		balancer: balancer,
	}

	as.registerMiddlewares()
	as.registerRoutes()
	return as
}

func (as ApiService) registerMiddlewares() {
	// Middlewares
	as.echo.HTTPErrorHandler = CustomHTTPErrorHandler

	as.echo.Use(middleware.Logger())
	as.echo.Use(middleware.Recover())
}

func (as ApiService) registerRoutes() {
	// Routes
	as.echo.GET("/services", as.getServices)
	as.echo.GET("/services/:service_name", as.getService)
	as.echo.POST("/services", as.addService)
	as.echo.DELETE("/services/:service_name", as.deleteService)
	as.echo.POST("/services/:service_name/destinations", as.addDestination)
	as.echo.DELETE("/services/:service_name/destinations/:destination_name", as.deleteDestination)
	as.echo.POST("/services/:service_name/check", as.addCheck)
	as.echo.DELETE("/services/:service_name/check", as.deleteCheck)
}

// Serve starts the api.
// Binds IP using HOST env or 0.0.0.0
// Binds to port using PORT env or 8000
func (as ApiService) Serve() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	address := host + ":" + port

	// Start server
	as.echo.Logger.Fatal(as.echo.Start(address))
}
