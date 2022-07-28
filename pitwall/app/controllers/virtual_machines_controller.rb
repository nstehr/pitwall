require 'bunny'
require "#{Rails.root}/lib/protos/vm_pb.rb"

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
        @virtualMachine = VirtualMachine.create(
            image: params[:image]
        )
        # TODO: extract to service
        exchange_name = "pitwall.orchestration"
        conn = Bunny.new.tap(&:start)
        ch = conn.create_channel
        exchange = ch.topic(exchange_name, :durable => true)
        req = ::Vm::CreateVMRequest.new(:imageName => params[:image])
        message = ::Vm::CreateVMRequest.encode(req)
        exchange.publish(message, routing_key: "orchestrator.vm.crud.leonardo")
        
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
