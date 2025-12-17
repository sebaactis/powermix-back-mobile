package coffeeji

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestValidateVoucherCode_DataHasInfo_ReturnsTrue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"code":200,"success":true,"data":{"orderId":"A1"},"msg":"ok"}`)
	}))
	defer ts.Close()

	key := os.Getenv("COFFEJI_KEY")
	secret := os.Getenv("COFFEJI_SECRET")

	c := &Client{baseURL: "https://gsvden.coffeeji.com", httpClient: ts.Client(), key: key, secret: secret}

	used, err := c.ValidateVoucherCode(context.Background(), "118595")

	fmt.Println(used)
	if err != nil {
		t.Fatal(err)
	}
	if used != true {
		t.Fatalf("expected true, got %v", used)
	}
}

func TestValidateVoucherCode_DataIsEmpty_ReturnsFalse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"code":200,"success":true,"data":{"orderId":"A1"},"msg":"ok"}`)
	}))
	defer ts.Close()

	key := os.Getenv("COFFEJI_KEY")
	secret := os.Getenv("COFFEJI_SECRET")

	c := &Client{baseURL: "https://gsvden.coffeeji.com", httpClient: ts.Client(), key: key, secret: secret}

	used, err := c.ValidateVoucherCode(context.Background(), "350585")

	fmt.Println(used)
	if err != nil {
		t.Fatal(err)
	}
	if used != false {
		t.Fatalf("expected false, got %v", used)
	}
}
