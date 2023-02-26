require 'petname'
class VirtualMachine < ApplicationRecord
    validates :image, presence: true
    validates :name, uniqueness: { scope: :user }
    belongs_to :orchestrator, counter_cache: true, optional: true
    belongs_to :user, counter_cache: true, optional: true
    after_initialize :set_default_name

    def set_default_name
        pn = PetName::Generator.new
        self.name = pn.generate if self.name.blank?
    end
end
