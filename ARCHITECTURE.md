```mermaid
C4Context
      title Pitwall Architecture

        Person(customerA, "CLI user")
        Person(customerB, "Web user")
        

        System(pitwall, "Pitwall", "Control plane.  Creating, deleting, monitoring microVMs")
        

        Boundary(ziti, "OpenZiti") {
            System(zRouter, "Router")
            System(zController, "Controller")
            System(zAC, "Admin Console UI")
          }

          Boundary(host, "Host") {
            System(vm, "microVM")
            System(firecracker, "Firecracker")
            System(orchestrator, "Orchestrator", "Communicates with the firecracker binary")
            System(terminator, "Terminator", "Proxies openziti connections to VMs")
          }

        
          SystemDb(postgres, "PostgresDB")
          SystemQueue(rabbit, "RabbitMQ")
          SystemDb(keycloak, "Keycloak")
          
          
        
      

      Rel(customerA, pitwall, "Uses")
      Rel(customerB, pitwall, "Uses")
      Rel(pitwall, postgres, "Uses")
      Rel(keycloak, postgres, "Uses")
      Rel(customerA, keycloak, "Authenticates against")
      Rel(customerB, keycloak, "Authenticates against")
      Rel(pitwall, keycloak, "Verifies with")
      Rel(orchestrator, firecracker, "manages")
      Rel(firecracker, vm, "manages")
      Rel(terminator, vm, "proxies")
      BiRel(customerA, zController, "")
      BiRel(terminator, zController, "")
      BiRel(terminator, rabbit, "")
      BiRel(orchestrator, rabbit, "")
      BiRel(pitwall, rabbit, "")


      UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="1")
```

# Overview
A pitwall deployment is made up of three main layers:
 - A web based control plane
 - Infrastructure
 - Host(s) for running the microVMs

# Component Communication
Inter-component communication is mostly asynchronous via rabbitMQ using a topic based exchange.