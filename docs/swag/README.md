# API Documentation (Swagger)

Itinera uses **Swaggo** to generate interactive API documentation based on code comments.

## Access
Once the backend is running, you can access the Swagger UI at:
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## Update Documentation
If you add new endpoints or modify models, you must regenerate the documentation files:

```bash
cd backend
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go
```

This will update the files in `backend/docs/`. Remember to commit these changes so the documentation stays up to date.

## Writing Documentation
Use declarative comments in your handlers. Example:

```go
// ListTrips godoc
// @Summary      List trips
// @Description  Get a list of trips associated with the current user or session
// @Tags         trips
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Trip
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /trips [get]
```
