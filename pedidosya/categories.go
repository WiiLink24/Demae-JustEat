package justeat

import "github.com/WiiLink24/DemaeJustEat/demae"

var categoryTypes = map[demae.CategoryCode][]string{
	demae.Pizza:            {"pizza", "italian-style-pizza"},
	demae.Western:          {"hamburger", "pollo", "americano", "messicano", "american", "sandwiches", "burger", "amerikanisches-essen"},
	demae.FastFood:         {"panini", "hamburger", "friti", "pollo", "chicken", "burger", "gefluegelgerichte", "doener"},
	demae.Chinese:          {"cinese", "asianfusion", "asian", "chinese", "asiatisch-essen"},
	demae.DrinksAndDessert: {"dolci", "gelato", "bevande", "desserts", "coffee", "cakes", "drinks", "getraenke-und-snacks", "snacks"},
	demae.Curry:            {"curry", "indian", "indian-food"},
	demae.Japanese:         {"asianfusion", "ramen", "japanese"},
	demae.PartyFood:        {"kebab", "doener"},
	demae.Sushi:            {"sushi"},
}
