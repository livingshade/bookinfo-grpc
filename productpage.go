package services

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/Romero027/bookinfo-grpc/proto/details"
	"github.com/Romero027/bookinfo-grpc/proto/reviews"
	"google.golang.org/grpc"
)

func (s *ProductPage) initializeProucts() {
	s.Products = []Product{{
		ID:    0,
		Title: "The Comedy of Errors",
	}, {
		ID:    1,
		Title: "1984",
	},
	}
}

func dial(addr string) *grpc.ClientConn {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		panic(fmt.Sprintf("ERROR: dial error: %v", err))
	}

	return conn
}

// NewFrontend returns a new server
func NewProductPage(port int, reviewsddr string, detailsaddr string) *ProductPage {
	return &ProductPage{
		port:          port,
		detailsClient: details.NewDetailsClient(dial(detailsaddr)),
		reviewsClient: reviews.NewReviewsClient(dial(reviewsddr)),
		User:          "None",
	}
}

// Frontend implements frontend service
type ProductPage struct {
	port          int
	detailsClient details.DetailsClient
	reviewsClient reviews.ReviewsClient
	User          string
	Products      []Product
}

type Product struct {
	ID      int
	Title   string
	Reviews []*reviews.Review
	Details []*details.Detail
}

// Run the server
func (s *ProductPage) Run() error {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/index", http.FileServer(http.Dir("static")))
	http.HandleFunc("/login", s.loginHandler)
	http.HandleFunc("/logout", s.logoutHandler)
	http.HandleFunc("/productpage", s.productpageHandler)
	http.HandleFunc("/products", s.productsHandler)
	http.HandleFunc("/reviews", s.reviewsHandler)
	http.HandleFunc("/details", s.detailsHandler)

	log.Printf("ProductPage server running at port: %d", s.port)
	s.initializeProucts()
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *ProductPage) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// user := r.URL.Query().Get("user")
	s.User = "Jason"

	json.NewEncoder(w).Encode("Login Success!")
}

func (s *ProductPage) logoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	s.User = "None"

	json.NewEncoder(w).Encode("Logout Success!")
}

func (s *ProductPage) productpageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()
	productID := 0
	reviewsRes, err := s.getReviewsDetails(ctx, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	detailRes, err := s.getProductDetails(ctx, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Products[productID].Reviews = reviewsRes.Review
	s.Products[productID].Details = detailRes.Detail

	tmpl := template.Must(template.ParseFiles("static/productpage.html"))

	tmpl.Execute(w, s)
}

func (s *ProductPage) productsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(s.Products)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (s *ProductPage) reviewsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()
	// productID, err := strconv.Atoi(r.URL.Query().Get("productID"))
	productID := 0

	reviewsRes, err := s.getReviewsDetails(ctx, productID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(reviewsRes)
}

func (s *ProductPage) detailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()
	// productID, err := strconv.Atoi(r.URL.Query().Get("productID"))
	productID := 0

	detailRes, err := s.getProductDetails(ctx, productID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(detailRes)
}

func (s *ProductPage) getProductDetails(ctx context.Context, id int) (*details.Result, error) {
	detailRes, err := s.detailsClient.GetDetails(ctx, &details.Product{
		Id: int32(id),
	})
	return detailRes, err
}

func (s *ProductPage) getReviewsDetails(ctx context.Context, id int) (*reviews.Result, error) {
	reviewsRes, err := s.reviewsClient.GetReviews(ctx, &reviews.Product{
		Id: int32(id),
	})
	return reviewsRes, err
}
