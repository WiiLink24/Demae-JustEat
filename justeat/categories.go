package justeat

import "github.com/WiiLink24/DemaeJustEat/demae"

var categoryTypes = map[demae.CategoryCode][]string{
	demae.Pizza:            {"pizza", "italian-style-pizza", "italienische-pizza", "amerikanische-pizza", "tuerkische-pizza", "pide", "pinsa", "flammkuchen"},
	demae.Western:          {"hamburger", "pollo", "americano", "messicano", "american", "sandwiches", "burger", "amerikanisches-essen", "italienisch", "italian-food", "pasta", "nudeln", "deutsches-essen", "deutsche-gerichte", "hausmannskost", "oesterreichische-kuche", "schnitzel", "griechisches-essen", "spanisch-und-tapas", "franzoesisches-essen", "portugiesisch-essen", "balkan", "polnisches-essen", "russisches-essen", "ukrainisches-essen", "mexikanisch", "lateinamerikanisch", "argentinisch-essen", "brasilianisch-essen", "steaks", "spare-ribs", "grillgerichte", "rindfleisch", "schweinefleisch", "kalb", "fisch", "meeresfruchte-essen", "suppen", "mittagessen", "internationales-essen", "sonstiges-essen", "fine_dining", "vegetarisches-essen", "vegan", "vegan-essen", "bioessen", "glutenfreies-essen"},
	demae.FastFood:         {"panini", "hamburger", "friti", "pollo", "chicken", "burger", "gefluegelgerichte", "doener", "pommes", "hot-dog", "wraps", "bagels", "franzosische-tacos", "haehnchen", "salat", "bowls"},
	demae.Chinese:          {"cinese", "asianfusion", "asian", "chinese", "asiatisch-essen", "chinesisch", "chinese-food", "sichuan", "dumplings", "fruhlingsrollen", "vietnamesisch", "thailaendisch", "indonesisches-essen", "koreanisches-essen", "koreanisch-essen"},
	demae.DrinksAndDessert: {"dolci", "gelato", "bevande", "desserts", "coffee", "cakes", "drinks", "getraenke-und-snacks", "snacks", "eiscreme", "torte", "donuts", "backwaren", "pfannkuchen", "bubble_tea", "alkohol", "bei-cafes", "fruehstueck"},
	demae.Curry:            {"curry", "indian", "indian-food", "indisch", "pakistanisches-essen", "afghanisch-essen", "persisches-essen"},
	demae.Japanese:         {"asianfusion", "ramen", "japanese", "japanisches-essen", "poke-bowl"},
	demae.PartyFood:        {"kebab", "doener", "tuerkisches-essen", "arabisches-essen", "libanesisches-essen", "syrische-gerichte", "israelisches-essen", "marokkanisches-essen", "afrikanisches-essen", "orientalisch", "falafel", "gyros", "100-prozent-halal", "halal", "koscher"},
	demae.Sushi:            {"sushi"},
}
