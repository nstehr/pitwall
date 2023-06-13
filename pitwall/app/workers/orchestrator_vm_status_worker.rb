require "#{Rails.root}/lib/protos/vm_pb.rb"
class OrchestratorVmStatusWorker
    include Sneakers::Worker
    from_queue "orchestrator.vm.status.web", env: nil
    # TODO: move the business logic out of here into a service
    def work(msg)
        vm = ::Vm::VM.decode(msg)
        virtual_machine = VirtualMachine.find(vm.id)
        services = Array.new
        if !vm.services.blank?
          vm.services.each do |service|
            s = {
              name: service.name,
              port: service.port,
              private: service.private,
              protocol: service.protocol, 
              virtual_machine_id: virtual_machine.id
            }
            services.append(s)
          end
          Service.insert_all(services)
        end
       
        virtual_machine.update(
            status: vm.status
        )
        if vm.status == "STOPPED"
            virtual_machine.update(
            orchestrator_id: nil 
        )
            virtual_machine.services.delete_all
        end
      ack! 
    end
  end