require 'bunny'
require "#{Rails.root}/lib/protos/vm_pb.rb"

class VirtualMachinePlacer
    def initialize(image)
        @image = image
    end
    def place()
        # this logic will change once the orchestrator is enriched
        # for now, select least used orchestrator
        orchestrators = Orchestrator.all
        orchestrator = orchestrators[0]
        orchestrators.each do |o| 
            if o.virtual_machines.count < orchestrator.virtual_machines.count
                orchestrator = o
        end

        vm = VirtualMachine.create(
            image: @image,
            orchestrator: orchestrator
        )
        exchange_name = "pitwall.orchestration"
        conn = Bunny.new.tap(&:start)
        ch = conn.create_channel
        exchange = ch.topic(exchange_name, :durable => true)
        req = ::Vm::CreateVMRequest.new(:imageName => @image)
        message = ::Vm::CreateVMRequest.encode(req)
        exchange.publish(message, routing_key: "orchestrator.vm.crud.#{orchestrator.name}")
        return vm
    end
end
end
