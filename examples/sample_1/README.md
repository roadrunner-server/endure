### Sample of Endure usage

In this sample we have logger interface which every plugin consumes (depends on). Therefore, it should be initialized first.
Database module depends on the logger and that's it. Next - http layer. It depends on the database and logger. The last in order are plugins for the http -> gzip and headers. Endure will resolve these dependencies in the following order:
1. Logger (2 deps, HTTP and DB)
2. Database (1 dep, logger)
3. Plugins (depends via Middleware interface on HTTP)
4. HTTP

Each module is independent and can be hidden via interface in the `Init` function.