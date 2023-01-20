module OrchestratorsHelper
        def get_orch_status_color(status)
           case status
           when "UP"
            "#65c3ba"
           when "DOWN"
            "#4a4e4d"
           else
            "#e6e6ea"
        end
    end
end
