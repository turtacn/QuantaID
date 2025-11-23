package quantaid.authz

# Default values
default allow = false
default deny = false

# Allow admin users access to everything
allow {
    input.user.roles[_] == "admin"
}

# Example: Allow view action on reports (can be extended)
# allow {
#     input.action == "view"
#     input.resource.type == "report"
# }
