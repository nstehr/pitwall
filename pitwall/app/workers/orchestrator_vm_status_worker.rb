require "#{Rails.root}/lib/protos/vm_pb.rb"
class OrchestratorVmStatusWorker
    include Sneakers::Worker
  
    from_queue "orchestrator.vm.status", env: nil
    def work(msg)
        vm = ::Vm::VM.decode(msg)
        virtualMachine = VirtualMachine.find(vm.id)
        virtualMachine.update(
            status: vm.status
        )
        if vm.status == "STOPPED"
            virtualMachine.update(
            orchestrator_id: nil 
        )
        end
      ack! 
    end
  end