class Api::VirtualMachinesController < Api::ApiController
    before_action :authenticate_user!
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
       

        vm = VirtualMachine.new(
            image: params[:image],
            public_key: params[:public_key],
            user: current_user,
            name: params[:name]
        )
        @virtual_machine = placer.place(vm)
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
