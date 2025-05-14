package main

import (
	"encoding/xml"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat"
)

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
