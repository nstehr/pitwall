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
        placer = VirtualMachinePlacer.new
        @virtualMachine = placer.place(params[:image])
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
        placer = VirtualMachinePlacer.new
        @virtualMachine = placer.stop(params[:id])
        render json: @virtualMachines
    end
end
