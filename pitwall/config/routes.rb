Rails.application.routes.draw do
  namespace :api do 
      resources :virtual_machines
      resources :orchestrators
  end
  resources :virtual_machines
end
