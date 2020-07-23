### Sample of cascade usage

Here we have folder with modules. Each module is independent. But, it should be run in order
1. Logger module  
2. DB module  
3. Http module
4. Gzip or Headers (0 deps)
5. Headers or Gzip (0 deps)