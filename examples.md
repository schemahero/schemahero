## show all migrations, sorted by age

kubectl schemahero get migrations -d replicated

## show all planned but not yet approved, executed, or rejected, sorted by age
kubectl schemahero get migrations -d replicated --status=planned

## show all executed but not rejected, sorted by age
kubectl schemahero get migrations -d replicated --status=executed

## show all executed but not rejected, sorted by age
kubectl schemahero get migrations -d replicated --status=approved

## show all rejected, sorted by age
kubectl schemahero get migrations -d replicated --status=rejected

## Example raw output from current `kubectl schemahero get migrations -d replicated`:

```
ID       DATABASE    TABLE                                             PLANNED   EXECUTED  APPROVED  REJECTED
f9bc491  replicated  cluster-history                                   805d21h   805d21h   805d21h   
f9f8d8e  replicated  app-release-preflight-checks                      1h                            
fa1e3a4  replicated  enterprise-user                                   1h                            
fb0f3c2  replicated  kots-license-instance                             544d19h   544d19h   544d19h   
fe0e877  replicated  entitlement-spec                                  450d2h                        
fe4f77f  replicated  kots-entitlement-field                            856d3h    856d2h    856d2h    
fe75c65  replicated  kots-licenses                                     449d19h   449d19h   449d19h   449d19h
```