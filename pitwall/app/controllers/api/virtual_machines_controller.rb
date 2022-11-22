class Api::VirtualMachinesController < Api::ApiController
    def index
        @virtual_machines = VirtualMachine.all 
        render json: @virtual_machines
    end 

    def show
        @virtual_machine = VirtualMachine.find(params[:id])
        render json: @virtual_machine
    end 

    def create
        placer = VirtualMachinePlacer.new
        @virtual_machine = placer.place(params[:image], params[:public_key])
        render json: @virtual_machine
    end 

    def update
        @virtual_machine = VirtualMachine.find(params[:id])
        @virtual_machine.update(
            image: params[:image]
        )
        render json: @virtual_machine
    end 

    def destroy
        placer = VirtualMachinePlacer.new
        @virtual_machine = placer.stop(params[:id])
        render json: @virtual_machines
    end
end
