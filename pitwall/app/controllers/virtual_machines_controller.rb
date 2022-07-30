class VirtualMachinesController < ApplicationController
    def index
        @virtualMachines = VirtualMachine.all 
        render json: @virtualMachines
    end 

    def show
        @virtualMachine = VirtualMachine.find(params[:id])
        render json: @virtualMachine
    end 

    def create
        placer = VirtualMachinePlacer.new(params[:image])
        @virtualMachine = placer.place
        render json: @virtualMachine
    end 

    def update
        @virtualMachine = VirtualMachine.find(params[:id])
        @virtualMachine.update(
            image: params[:image]
        )
        render json: @virtualMachine
    end 

    def destroy
        @virtualMachines = VirtualMachine.all 
        @virtualMachine = VirtualMachine.find(params[:id])
        @virtualMachine.destroy
        render json: @virtualMachines
    end
end
