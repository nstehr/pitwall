require "#{Rails.root}/lib/protos/vm_pb.rb"
class OrchestratorVmStatusWorker
    include Sneakers::Worker
    from_queue "orchestrator.vm.status.web", env: nil
    def work(msg)
        vm = ::Vm::VM.decode(msg)
        virtual_machine = VirtualMachine.find(vm.id)
        services = Array.new
        if !vm.services.blank?
          vm.services.each do |service|
            s = Service.new
            s.name = service.name
            s.port = service.port
            s.private = service.private
            s.protocol = service.protocol
            s.virtual_machine = virtual_machine
            services.append(s)
          end
      end
  
        virtual_machine.update(
            status: vm.status,
            services: services
        )
        if vm.status == "STOPPED"
            virtual_machine.update(
            orchestrator_id: nil 
        )
        end
      ack! 
    end
  end