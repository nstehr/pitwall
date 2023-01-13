class OrchestratorsController < ApplicationController
    before_action :authenticate_user!
    def index
        @orchestrators = Orchestrator.all.order(created_at: :desc)
    end 
end
