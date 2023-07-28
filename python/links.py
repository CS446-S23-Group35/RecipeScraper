cats = ["l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "xyz"]
totalPages = {
    "l": 16,
    "m": 31,
    "n": 7,
    "o": 10,
    "p": 45,
    "q": 3,
    "r": 27,
    "s": 77,
    "t": 28,
    "u": 2,
    "v": 7,
    "w": 13,
    "xyz": 4
}

with open("links.txt", "w") as f:
    for cat in cats:
        for i in range(1, totalPages[cat] + 1):
            # https://www.foodnetwork.com/recipes/recipes-a-z/w/p/2
            link = f"https://www.foodnetwork.com/recipes/recipes-a-z/{cat}"
            if i != 1:
                link += f"/p/{i}"
            link += "\n"
            f.write(link)