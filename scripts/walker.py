import os

def walk_directory(path):
    tree = {}

    for item in os.listdir(path):
        full_path = os.path.join(path, item)

        if os.path.isdir(full_path):
            tree[item] = walk_directory(full_path)
        else:
            tree[item] = {}

    return tree


if __name__ == "__main__":
    path = input("Enter directory path: ")

    result = walk_directory(path)
    print(result)
