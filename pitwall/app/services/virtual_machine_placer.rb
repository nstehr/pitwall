require "#{Rails.root}/lib/protos/vm_pb.rb"

class VirtualMachinePlacer
    def place(image)
        # this logic will change once the orchestrator is enriched
        # for now, select least used orchestrator
        orchestrators = Orchestrator.all
        orchestrator = orchestrators[0]
        orchestrators.each do |o|
            if o.virtual_machines.count < orchestrator.virtual_machines.count
                orchestrator = o
            end
        end
           
        vm = VirtualMachine.create(
            image: image,
            orchestrator: orchestrator,
            status: "INIT"
        )
       
        req = ::Vm::CreateVMRequest.new(:id => vm.id, :imageName => image)
        message = ::Vm::CreateVMRequest.encode(req)
        routing_key = "orchestrator.vm.crud.#{orchestrator.name}"
        rabbit = Rabbitmq.new()
        rabbit.send(routing_key, message)
        return vm
    end
end
