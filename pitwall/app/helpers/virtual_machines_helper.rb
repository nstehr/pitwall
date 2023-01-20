module VirtualMachinesHelper
    def get_status_color(status)
       case status
       when "RUNNING"
        "#65c3ba"
       when "INIT"
        "#e6e6ea"
       when "STOPPED"
        "#4a4e4d"
       when "STOPPING"
        "#4a4e4d"
       when "ERROR"
        "#fe4a49"
       when "BOOTING"
        "#8874a3"
       when "BUILDING_FILESYSTEM"
        "#8874a3"
       else
        "#e6e6ea"
    end
end
end
