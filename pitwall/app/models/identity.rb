class Identity < ApplicationRecord
  belongs_to :user
  validates :public_key, public_key: true
end
