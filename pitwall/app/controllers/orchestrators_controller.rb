class OrchestratorsController < ApplicationController
    def index
        @orchestrators = Orchestrator.all 
        render json: @orchestrators
    end 

    def show
        @orchestrator = Orchestrator.find(params[:id])
        render json: @orchestrator
    end 

    def create
        @orchestrator = Orchestrator.create(
            name: params[:name],
            status: params[:status]
        )
        render json: @orchestrator
    end 

    def update
        @orchestrator = Orchestrator.find(params[:id])
        @orchestrator.update(
            name: params[:name],
            status: params[:status]
        )
        render json: @orchestrator
    end 

    def destroy
        @orchestrators = Orchestrator.all 
        @orchestrator = Orchestrator.find(params[:id])
        @orchestrator.destroy
        render json: @rchestrators
    end 
end
