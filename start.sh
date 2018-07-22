
floe -tags=linux,go,couch -admin=123456 -host_name=h1 -pub_bind=127.0.0.1:8080 >> h1.log 2>&1 &
floe -tags=linux,go,couch -admin=123456 -host_name=h2 -pub_bind=127.0.0.1:8090 >> h2.log     2>&1&


