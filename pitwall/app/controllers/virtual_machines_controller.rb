class VirtualMachinesController < ApplicationController
    before_action :authenticate_user!
    def index
        if params[:status].present?
           @pagy, vm_list = pagy(VirtualMachine.where(status: params[:status]).order(created_at: :desc))
        else
           @pagy, vm_list = pagy(VirtualMachine.all.order(created_at: :desc))
        end
        @virtual_machines = vm_list
    end 

    def show
        @virtual_machine = VirtualMachine.find(params[:id])
        render json: @virtual_machine
    end

    def new
        @virtual_machine = VirtualMachine.new
    end

    def create
        placer = VirtualMachinePlacer.new
        image = params[:virtual_machine][:image]
        public_key = params[:virtual_machine][:public_key]
        name = params[:virtual_machine][:name]

        vm = VirtualMachine.new(
            image: image,
            public_key: public_key,
            user: current_user,
            name: name
        )

        @virtual_machine = placer.place(vm)
        if @virtual_machine.valid?
            redirect_to virtual_machines_path, notice: "VM created successfully"
        else
            render :new, status: :unprocessable_entity
        end
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
