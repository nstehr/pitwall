require "#{Rails.root}/lib/protos/vm_pb.rb"

class VirtualMachinePlacer
    def place(vm)
        # this logic will change once the orchestrator is enriched
        # for now, select least used orchestrator
        orchestrators = Orchestrator.all
        orchestrator = orchestrators[0]
        orchestrators.each do |o|
            if o.virtual_machines.count < orchestrator.virtual_machines.count
                orchestrator = o
            end
        end
       
        vm.orchestrator = orchestrator
        vm.status =  "INIT"
       
        saved = vm.save

        if saved
            create = ::Vm::CreateVMRequest.new(:id => vm.id, :imageName => vm.image, :publicKey => vm.public_key)
            req = ::Vm::VMRequest.new(:type => ::Vm::Type::CREATE, :create => create)
            message = ::Vm::VMRequest.encode(req)
            routing_key = "orchestrator.vm.crud.#{orchestrator.name}"
            rabbit = Rabbitmq.new()
            rabbit.send(routing_key, message)
        end
        return vm
    end

    def stop(id)
        vm = VirtualMachine.find(id)
        vm.update(
            status: "STOPPING"
        )
        orchestrator = vm.orchestrator
        stop = ::Vm::StopVMRequest.new(:id => vm.id)
        req = ::Vm::VMRequest.new(:type => ::Vm::Type::DELETE, :stop => stop)
        message = ::Vm::VMRequest.encode(req)
        routing_key = "orchestrator.vm.crud.#{orchestrator.name}"
        rabbit = Rabbitmq.new()
        rabbit.send(routing_key, message)
        return vm
    end
end
