package quantaid.authz

default allow = false

# Allow administrators to do anything
allow {
    input.user.roles[_] == "admin"
}

# Allow users to read their own profile
allow {
    input.action == "read"
    input.resource.type == "user_profile"
    input.resource.id == input.user.id
}

# Example of ABAC: Allow write only during work hours (simulated)
# In a real scenario, you'd pass the current time in the input or use OPA's time functions
allow {
    input.action == "write"
    is_work_hours
}

is_work_hours {
    # This is a placeholder. In production, you might check input.env.time or use time.now_ns()
    # For now, we assume it's always work hours for demonstration,
    # or you can enforce it via input variables.
    true
}
