package justeat

import (
	"context"
	"testing"
)

func TestOrder(t *testing.T) {
	ctx := context.Background()

	pool := initDB(t, ctx)
	defer pool.Close()

	req := makeFakeRequest()
	client, err := NewClient(ctx, pool, req, HollywoodID)
	if err != nil {
		panic(err)
	}

	// We are going to test with KFC.
	restID := "kfc-sevensisters"
	// Burgers
	// categoryID := "abdfba10-5680-4e1a-9ea5-f5ed8fae96e1"
	// Zinger Supercharger Tower Burger i think
	itemID := "f7a2638b-b878-5a9e-a5f7-f45fc13f926f"

	// Set up the POST form to parse products.
	req.PostForm.Set("itemCode", itemID)
	req.PostForm.Set("shopCode", restID)
	req.PostForm.Set("quantity", "1")

	_, err = client.CreateBasket(req)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetAvailableTimes("NGFkNjVjMTYtMWY4MS00ZG-v1")
	if err != nil {
		t.Fatal(err)
	}
}
