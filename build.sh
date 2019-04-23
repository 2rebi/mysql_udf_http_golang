version_info=$(mysql --version)
if [[ "$version_info" == *"Maria"* ]]; then
    include_dir=$(mariadb_config --include)
else
    include_dir=$(mysql_config --include)
fi

export CGO_CFLAGS=$include_dir
go build -buildmode=c-shared -o http.so http.go
