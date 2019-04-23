if [[ $# == 2 || $# == 1 ]]; then
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION http_help;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION http_raw;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION http_get;"
    mysql --user=$1 --password=$2 -s -N -e "DROP FUNCTION http_post;"
    
    sql_result=$(mysql --user=$1 --password=$2 -s -N -e "SHOW VARIABLES LIKE 'plugin_dir';")
    plugin_dir=$(cut -d" " -f2 <<< $sql_result)
    rm $plugin_dir"http.so"

    echo "Uninstall Success"
else
    echo "bash uninstall.sh username password(optional)"
fi

