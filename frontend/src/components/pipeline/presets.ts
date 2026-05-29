export const PIPELINE_PRESETS = [
    {
        name: "Velkomstpakken",
        description: "Sender velkomst-epost til nye medlemmer automatisk.",
        nodes: [
            {
                id: "trigger",
                type: "trigger",
                position: { x: 50, y: 150 },
                data: { event: "MembershipCreated", outputs: ["user_id", "org_id", "role", "timestamp"] }
            },
            {
                id: "fmt_welcome",
                type: "action",
                position: { x: 400, y: 150 },
                dimensions: { width: 200, height: 150 },
                data: { 
                    label: "FormatText", 
                    inputs: ["template", "user_name"], 
                    outputs: ["result"],
                    template: "Hei {{user_name}}! Velkommen som medlem hos oss." 
                }
            },
            {
                id: "send_mail",
                type: "action",
                position: { x: 750, y: 150 },
                dimensions: { width: 200, height: 150 },
                data: { 
                    label: "SendEmail", 
                    inputs: ["to_email", "subject", "body"], 
                    outputs: ["success"],
                    subject: "Velkommen til oss!"
                }
            }
        ],
        edges: [
            { id: "e1", source: "trigger", sourceHandle: "timestamp", target: "fmt_welcome", targetHandle: "template", type: "smoothstep", animated: true },
            { id: "e2", source: "fmt_welcome", sourceHandle: "result", target: "send_mail", targetHandle: "body", type: "smoothstep", animated: true }
        ]
    },
    {
        name: "Purring på ubetalt kontingent",
        description: "Sjekker ukentlig og sender purring til medlemmer med negativ saldo.",
        nodes: [
            {
                id: "trigger",
                type: "trigger",
                position: { x: 50, y: 200 },
                data: { event: "TimedSchedule", aggregateId: "weekly", outputs: ["timestamp", "source"] }
            },
            {
                id: "find_members",
                type: "action",
                position: { x: 350, y: 200 },
                dimensions: { width: 200, height: 150 },
                data: { label: "FindMemberByRole", inputs: ["org_id", "role_name"], outputs: ["email", "name", "user_id"], role_name: "Member" }
            },
            {
                id: "check_balance",
                type: "logic",
                position: { x: 650, y: 200 },
                dimensions: { width: 240, height: 200 },
                data: { type: "if", operator: "<", value1: "0", value2: "0", inputs: ["v1", "v2", "operator"] }
            },
            {
                id: "send_reminder",
                type: "action",
                position: { x: 1000, y: 100 },
                dimensions: { width: 200, height: 150 },
                data: { label: "SendEmail", inputs: ["to_email", "subject", "body"], outputs: ["success"], subject: "Påminnelse: Ubetalt medlemskontingent" }
            }
        ],
        edges: [
            { id: "e1", source: "trigger", target: "find_members", type: "smoothstep", animated: true },
            { id: "e2", source: "find_members", target: "check_balance", type: "smoothstep", animated: true },
            { id: "e3", source: "check_balance", sourceHandle: "true", target: "send_reminder", type: "smoothstep", animated: true }
        ]
    }
]
