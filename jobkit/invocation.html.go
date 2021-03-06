package jobkit

var invocationTemplate = `
{{ define "invocation" }}
{{ template "header" . }}
<div class="container">
	<ul class="breadcrumbs">
		<li><a href="/">Jobs</a></li>
		<li>{{ .ViewModel.JobName }}</li>
		<li>{{ .ViewModel.ID }}</li>
	</ul>
	<table class="u-full-width">
		<thead>
			<tr>
				<th>Invocation</th>
				<th>Started</th>
				<th>Finished</th>
				<th>Timeout</th>
				<th>Cancelled</th>
				<th>Elapsed</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>{{ .ViewModel.ID }}</td>
				<td>{{ .ViewModel.Started | rfc3339 }}</td>
				<td>{{ if .ViewModel.Finished.IsZero }}-{{ else }}{{ .ViewModel.Finished | rfc3339 }}{{ end }}</td>
				<td>{{ if .ViewModel.Timeout.IsZero }}-{{ else }}{{ .ViewModel.Timeout | rfc3339 }}{{ end }}</td>
				<td>{{ if .ViewModel.Cancelled.IsZero }}-{{ else }}{{ .ViewModel.Cancelled | rfc3339 }}{{ end }}</td>
				<td>{{ .ViewModel.Elapsed }}</td>
			</tr>
		</tbody>
	</table>
	{{ if .ViewModel.Err }}
	<table class="u-full-width">
		<thead>
			<tr>
				<th>Error</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>
					<pre>{{ .ViewModel.Err }}</pre>
				</td>
			</tr>
		</tbody>
	</table>
	{{ end }}
	{{ if .ViewModel.State }}
	<table class="u-full-width">
		<thead>
			<tr>
				<th>Error Output</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>
					<pre>{{ .ViewModel.State.ErrorOutput }}</pre>
				</td>
			</tr>
		</tbody>
	</table>
	<table class="u-full-width">
		<thead>
			<tr>
				<th>Output</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>
					<pre>{{ .ViewModel.State.Output }}</pre>
				</td>
			</tr>
		</tbody>
	</table>
	{{ end }}
</div>
{{ template "footer" . }}
{{ end }}
`
