package main

import (
	"encoding/xml"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/justeat"
)

func menuList(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	menus, err := client.GetMenuCategories(r.request.URL.Query().Get("shopCode"))
	if err != nil {
		r.ReportError(err)
		return
	}

	// Append 1 more as a placeholder
	placeholder := menus[0]
	placeholder.XMLName = xml.Name{Local: "placeholder"}
	menus = append(menus, placeholder)
	r.AddCustomType(menus)
}

func itemList(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	items, err := client.GetMenuItems(r.request.URL.Query().Get("shopCode"), r.request.URL.Query().Get("menuCode"))
	if err != nil {
		r.ReportError(err)
		return
	}

	r.ResponseFields = []any{
		demae.KVField{
			XMLName: xml.Name{Local: "Count"},
			Value:   len(items),
		},
		demae.KVFieldWChildren{
			XMLName: xml.Name{Local: "List"},
			Value:   []any{items[:]},
		},
	}
}

func itemOne(r *Response) {
	client, err := justeat.NewClient(ctx, pool, r.request, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	itemCode := r.request.URL.Query().Get("itemCode")
	item, price, err := client.GetItemData(r.request.URL.Query().Get("shopCode"), r.request.URL.Query().Get("menuCode"), itemCode)
	if err != nil {
		r.ReportError(err)
		return
	}

	r.ResponseFields = []any{
		demae.KVField{
			XMLName: xml.Name{Local: "price"},
			Value:   price,
		},
		demae.KVFieldWChildren{
			XMLName: xml.Name{Local: "optionList"},
			Value:   []any{item[:]},
		},
	}
}
