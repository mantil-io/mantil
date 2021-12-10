
# mantil aws resources

Shows AWS resources created by Mantil

When executed inside Mantil project command will show resources created
for current project stage and node of that stage.
To show resources for other, non current, stage use --stage option.

When executed outside of Mantil project command will show resources of
the all nodes in the workspace.
Use --nodes options to get this behavior when inside of Mantil project.

### USAGE
<pre>
  mantil aws resources [options]
</pre>
### OPTIONS
<pre>
  -n, --nodes          Show resources for each workspace node
  -s, --stage string   Show resources for this stage
</pre>
### GLOBAL OPTIONS
<pre>
      --help       Show command help
      --no-color   Don't use colors in output
</pre>
### LEARN MORE
<pre>
  Visit https://github.com/mantil-io/docs to learn more.
  For further support contact us at support@mantil.com.
</pre>
