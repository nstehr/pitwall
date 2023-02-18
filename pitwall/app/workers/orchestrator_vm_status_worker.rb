require "#{Rails.root}/lib/protos/vm_pb.rb"
class OrchestratorVmStatusWorker
    include Sneakers::Worker
  
    from_queue "orchestrator.vm.status.web", env: nil
    def work(msg)
        vm = ::Vm::VM.decode(msg)
        virtual_machine = VirtualMachine.find(vm.id)
        virtual_machine.update(
            status: vm.status
        )
        if vm.status == "STOPPED"
            virtual_machine.update(
            orchestrator_id: nil 
        )
        end
      ack! 
    end
  end