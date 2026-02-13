import os

def walk_directory(path):
    tree = {}

    try:
        items = os.listdir(path)
    except Exception as e:
        print("Error accessing directory:", e)
        return {}

    for item in items:
        full_path = os.path.join(path, item)

        if os.path.isdir(full_path):
            try:
                tree[item] = walk_directory(full_path)
            except PermissionError:
                tree[item] = {"Permission Denied": {}}
        else:
            tree[item] = {}

    return tree


def print_tree(tree, indent=""):
    for key, value in tree.items():
        print(indent + "├── " + key)
        print_tree(value, indent + "│   ")


if __name__ == "__main__":
    path = input("Enter directory path: ")

    result = walk_directory(path)
    print_tree(result)
