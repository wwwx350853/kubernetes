##Command between CCE kubectl and community edition##

Command|Parameter|Available in CCE kubectl|
----- | ----- | ----- |
 get 	| componentstatuses | N |
 		| configmaps | N |
  		| daemonsets | N |
  		| deployments | N |
  		| events | N |
 		| endpoints | Y |
 		| horizontalpodautoscalers | N |
 		| ingress | N |
 		| jobs | N |
 		| limitranges | N |
 		| nodes | N |
     	| namespaces | Y |  
     	| pods | Y |
     	| persistentvolumes | N |
     	| persistentvolumeclaims | N |
     	| quota / resourcequotas | N |
     	| replicasets | N |
     	| replicationcontrollers | Y |
     	| secrets | N |
     	| secret {id} | Y |
     	| serviceaccounts | N |
     	| services | Y |
create	|endpoints | Y |
	  	|namespaces | Y |
		|pods | Y |
		|replicationcontrollers | Y |
		|services | Y |
		|namespaces | Y |		
replace |	endpoints | Y |
		|namespaces | Y |	
		|pods | Y |
		|replicationcontrollers	| Y |
		|secrets | Y |
		|services | Y |	
delete	|endpoints | Y |
		|namespaces | Y |	
		|pods | Y |
		|replicationcontrollers	| Y |
		|secrets | Y |
		|services | Y |	
convert	|	| Y |
patch 	|endpoints	| Y |
		|namespaces| Y |	
		|pods	| Y |
		|replicationcontrollers	| Y |
		|services | Y |
expose  |pods	| Y |
		|replicationcontrollers	| Y |
		|services	| Y |
annotate|endpoints	| Y |
		|namespaces | Y |	
		|pods	| Y |
		|replicationcontrollers	| Y |
		|services	| Y |
label	|endpoints	| Y |
		|namespaces| Y |	
		|pods	| Y |
		|replicationcontrollers	| Y |
		|services	| Y |
cluster-info|	| Y |
logs	|| Y |
api-version|| Y |
version || Y |
config  || Y |
apply   || Y |
rolling-update || Y |
scale || Y |
proxy || Y |
run   || Y |
		
		
		

