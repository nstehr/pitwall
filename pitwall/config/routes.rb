Rails.application.routes.draw do
  root to: 'home#index'
  Devise.setup do |config|
    devise_for :users, controllers: { omniauth_callbacks: 'users/omniauth_callbacks' }
  end
  devise_scope :user do
    get 'sign_in', :to => 'devise/sessions#new', :as => :new_user_session
    get 'sign_out', :to => 'devise/sessions#destroy', :as => :destroy_user_session
  end
  namespace :api do 
      resources :virtual_machines
      resources :orchestrators
      post 'profile/ztIdentity', to: 'users/profile#create_zt_identity'
  end
  resources :virtual_machines
  resources :orchestrators
  get 'profile', to: 'users/profile#show'
  get 'profile/public_key', to: 'users/profile#show_public_key'
  post 'profile/public_key', to: 'users/profile#update_public_key'
  patch 'profile/public_key', to: 'users/profile#update_public_key'
end