import yaml
import random

file = "recipes/recipes.yaml"
with open(file, 'r') as stream:
    try:
        data = yaml.load(stream, Loader=yaml.FullLoader)
        print(len(data))
        for recipe in random.sample(data, 100):
            with open("recipes/recipes/" + recipe["name"].replace(" ", "_").lower() + "_in.yaml", 'w') as outfile:
                yaml.dump(recipe["ingredientdescriptions"], outfile, default_flow_style=False)
    except yaml.YAMLError as exc:
        print(exc)
