package vm

type ApiServer struct {
	UnimplementedVMServiceServer
}

func NewApiServer() *ApiServer {
	return &ApiServer{}
}
func (as *ApiServer) CreateVM(in *CreateVMRequest, stream VMService_CreateVMServer) error {
	vm := VM{}
	vm.Status = "Pending"
	stream.Send(&vm)
	vm.Status = "Building Filesystem"
	stream.Send(&vm)
	buildFilesystemFromImage(in.GetImageName())
	vm.Status = "Up"
	stream.Send(&vm)
	return nil
}
