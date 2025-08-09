# TODO: @shravan20 - Delte before merging to master or feat/dev
echo "itr \tBinary RUN\t Poetry Run"

x=()
y=()


for i in {1..3}; do
    x[$i]=$( (time ./dist/nixopus --help > /dev/null 2>&1) 2>&1 | grep real | awk '{print $2}' )
    echo  "$i\t${x[$i]}\t\t-"
done

for i in {1..3}; do
    y[$i]=$( (time poetry run nixopus --help > /dev/null 2>&1) 2>&1 | grep real | awk '{print $2}' )
    echo "$i\t-\t\t   ${y[$i]}"
done