# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: vm.proto

require 'google/protobuf'

Google::Protobuf::DescriptorPool.generated_pool.build do
  add_file("vm.proto", :syntax => :proto3) do
    add_message "vm.VMRequest" do
      optional :type, :enum, 1, "vm.Type"
      oneof :payload do
        optional :create, :message, 2, "vm.CreateVMRequest"
        optional :stop, :message, 3, "vm.StopVMRequest"
      end
    end
    add_message "vm.CreateVMRequest" do
      optional :id, :int64, 1
      optional :imageName, :string, 2
    end
    add_message "vm.StopVMRequest" do
      optional :id, :int64, 1
    end
    add_message "vm.VM" do
      optional :id, :int64, 1
      optional :imageName, :string, 2
      optional :status, :string, 3
    end
    add_enum "vm.Type" do
      value :CREATE, 0
      value :DELETE, 1
    end
  end
end

module Vm
  VMRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("vm.VMRequest").msgclass
  CreateVMRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("vm.CreateVMRequest").msgclass
  StopVMRequest = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("vm.StopVMRequest").msgclass
  VM = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("vm.VM").msgclass
  Type = ::Google::Protobuf::DescriptorPool.generated_pool.lookup("vm.Type").enummodule
end
