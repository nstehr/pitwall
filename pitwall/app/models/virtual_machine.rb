class VirtualMachine < ApplicationRecord
    belongs_to :orchestrator, counter_cache: true, optional: true
end
