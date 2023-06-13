require 'petname'
class VirtualMachine < ApplicationRecord
    validates :image, presence: true
    validates :name, uniqueness: { scope: :user }
    validates :public_key, public_key: true
    belongs_to :orchestrator, counter_cache: true, optional: true
    belongs_to :user, counter_cache: true, optional: true
    has_many :services, dependent: :destroy
    after_initialize :set_default_name

    scope :by_user, ->(user) {where('user_id = ?', user.id)}

    def set_default_name
        pn = PetName::Generator.new
        self.name = pn.generate if self.name.blank?
    end
end
