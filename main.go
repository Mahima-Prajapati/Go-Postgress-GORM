package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Mahima-Prajapati/Go-Postgress-GORM/models"
	"github.com/Mahima-Prajapati/Go-Postgress-GORM/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	body := models.Books{}
	err := context.BodyParser(&body)
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{"error": "request failed"})
		return err
	}

	err = r.DB.Create(&body).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": "couldn't create book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "book created successfully"})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	book := models.Books{}

	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"error": "id cannot be empty"})
		return nil
	}

	err := r.DB.Delete(book, id).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": "couldn't delete book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "book deleted successfully"})
	return nil
}

func (r *Repository) BooksById(context *fiber.Ctx) error {
	book := &models.Books{}

	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"error": "id cannot be empty"})
		return nil
	}

	err := r.DB.Where("id = ?", id).First(book).Error
	if err != nil {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"error": "couldn't fetched book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book fetched successfully",
		"data":    book,
	})
	return nil
}

func (r *Repository) Books(context *fiber.Ctx) error {
	books := &[]models.Books{}

	err := r.DB.Find(books).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"error": "couldn't get books"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "books fetched successfully",
		"data":    books,
	})
	return nil
}

func (r *Repository) SetUpRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_book", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.BooksById)
	api.Get("/books", r.Books)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("couldn't connect to db")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("couldn't migrate db")
	}

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetUpRoutes(app)
	app.Listen(":8080")
}
