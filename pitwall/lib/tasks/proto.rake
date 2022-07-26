
# still learning rails and module learning.
# for now, dropping generated code in lib/protos based on
# here: https://github.com/t0yohei/research-rails7-nuxt3-protobuf/blob/537d9b1c07bb4dbb81d528f7e4e276a29da6d99d/backend/app/controllers/apiv2/todo_controller.rb#L1
# since it is the only approach I could get to work for now :) 
namespace :proto do 
    task :generate do
        sh "protoc --proto_path=../orchestrator/proto --ruby_out=lib/protos ../orchestrator/proto/orchestrator.proto"
    end
end
