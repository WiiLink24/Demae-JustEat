package main

import (
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat"
)

func shopInfo(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	categories, err := client.GetBareRestaurants()
	if err != nil {
		r.ReportError(err)
		return
	}

	// For some reason, the function that allows for reordering food skips the first two categories.
	shops := make([]demae.KVFieldWChildren, 2)
	for _, category := range categories {
		restaurants, err := client.GetRestaurants(category)
		if err != nil {
			r.ReportError(err)
			return
		}

		var shortenedRestaurants []demae.AreaShopInfo
		for i, restaurant := range restaurants {
			shortenedRestaurants = append(shortenedRestaurants, demae.AreaShopInfo{
				XMLName:  xml.Name{Local: fmt.Sprintf("container%d", i)},
				ShopCode: restaurant.ShopCode,
			})
		}

		container := demae.KVFieldWChildren{
			XMLName: xml.Name{Local: "container"},
			Value: []any{
				demae.KVField{
					XMLName: xml.Name{Local: "CategoryCode"},
					Value:   category,
				},
				demae.KVFieldWChildren{
					XMLName: xml.Name{Local: "ShopList"},
					Value: []any{
						shortenedRestaurants,
					},
				},
			},
		}

		shops = append(shops, container)

	}

	r.AddCustomType(shops)
}

func shopList(r *Response) {
	categoryCode := r.request.URL.Query().Get("categoryCode")

	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	restaurants, err := client.GetRestaurants(demae.CategoryCode(categoryCode))
	if err != nil {
		r.ReportError(err)
		return
	}

	shops := demae.KVFieldWChildren{
		XMLName: xml.Name{Local: "Pizza"},
		Value: []any{
			demae.KVField{
				XMLName: xml.Name{Local: "LargeCategoryName"},
				Value:   "Meal",
			},
			demae.KVFieldWChildren{
				XMLName: xml.Name{Local: "CategoryList"},
				Value: []any{
					demae.KVFieldWChildren{
						XMLName: xml.Name{Local: "TestingCategory"},
						Value: []any{
							demae.KVField{
								XMLName: xml.Name{Local: "CategoryCode"},
								Value:   categoryCode,
							},
							demae.KVFieldWChildren{
								XMLName: xml.Name{Local: "ShopList"},
								Value: []any{
									restaurants,
								},
							},
						},
					},
				},
			},
		},
	}

	r.AddCustomType(shops)
}

func shopOne(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	restaurant, err := client.GetRestaurant(r.request.URL.Query().Get("shopCode"))
	if err != nil {
		r.ReportError(err)
		return
	}

	r.ResponseFields = *restaurant
}
