class VirtualMachine < ApplicationRecord
    validates :image, presence: true
    belongs_to :orchestrator, counter_cache: true, optional: true
end
