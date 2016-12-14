##Command between CCE kubectl and community edition##

Command|Parameter|Available in CCE kubectl|Description|
----- | ----- | ----- | ----- |
 get 	| endpoints | Y |
     	| namespaces | Y |  
     	| pods | Y |
     	| replicationcontrollers | Y |
     	| secrets | Y |Only one secret can be queried at a time. For example, the command for querying secret -a is Kubectl get secret secret-a.
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
		
		
		

