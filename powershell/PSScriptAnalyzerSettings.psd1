@{
    # Comment-based help is optional for this thin argv wrapper; CI still fails
    # Error and Warning findings (automatic $args, missing UTF-8 BOM, etc.).
    ExcludeRules = @(
        'PSProvideCommentHelp'
    )
}
