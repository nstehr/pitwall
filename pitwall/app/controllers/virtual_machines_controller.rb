class VirtualMachinesController < ApplicationController
    def index
        if params[:status].present?
           @virtualMachines = VirtualMachine.where(status: params[:status]).order(created_at: :desc)
        else
           @virtualMachines = VirtualMachine.all.order(created_at: :desc)
        end
    end 

    def show
        @virtualMachine = VirtualMachine.find(params[:id])
        render json: @virtualMachine
    end

    def new
        @virtualMachine = VirtualMachine.new
    end

    def create
        placer = VirtualMachinePlacer.new
        image = params[:virtual_machine][:image]
        public_key = params[:virtual_machine][:public_key]
        @virtualMachine = placer.place(image, public_key)
        if @virtualMachine.valid?
            redirect_to virtual_machines_path, notice: "VM created successfully"
        else
            render :new, status: :unprocessable_entity
        end
    end 

    def update
        @virtualMachine = VirtualMachine.find(params[:id])
        @virtualMachine.update(
            image: params[:image]
        )
        render json: @virtualMachine
    end 

    def destroy
        placer = VirtualMachinePlacer.new
        @virtualMachine = placer.stop(params[:id])
        render json: @virtualMachines
    end
end
