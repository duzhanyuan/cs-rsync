<div class="container-fluid">
	<div class="row-fluid">
		<div class="hero-unit">
			<table class="table">
				<thead>
					<tr>
						<th>远程文件夹</th>
						<th>本地文件夹</th>
						<th>下载</th>
						<th>上传</th>
						<th>是否同步删除</th>
						<th>下载更新时间</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{{range .SyncInfoSlice}}
					<tr>
						<td>{{.RemoteDir}}</td>
						<td>{{.LocalDir}}</td>
						<td>{{.IsLoad}}</td>
						<td>{{.IsUpload}}</td>
						<td>{{.IsDelete}}</td>
						<td>{{.LoadUpdateTime}}</td>
						<td><button class="btn" type="button"><a href="/SyncInfo?Operate=Delete&RemoteDir={{.RemoteDir}}&LocalDir={{.LocalDir}}&IsLoad={{.IsLoad}}&IsUpload={{.IsUpload}}&IsDelete={{.IsDelete}}&LoadUpdateTime={{.LoadUpdateTime}}" >删除</a></button></td>
					</tr>
					{{end}}
				</tbody>
			</table>
		</div>
		<div class="span12">
			<form class="form-horizontal" method="POST" action="/SyncInfo?Operate=Add">
				<div class="control-group">
					<label class="control-label" for="RemoteDir">远程文件夹</label>
					<div class="controls"><input id="RemoteDir" name="RemoteDir" type="text" /></div>
				</div>
				<div class="control-group">
					<label class="control-label" for="LocalDir">本地文件夹</label>
					<div class="controls"><input id="LocalDir" name="LocalDir" type="text" /></div>
				</div>
				<div class="control-group">
					<label class="control-label" for="IsLoad">上传</label>
					<div class="controls"><input id="IsLoad" name="IsLoad" type="text" /></div>
				</div>
				<div class="control-group">
					<label class="control-label" for="IsUpload">下载</label>
					<div class="controls"><input id="IsUpload" name="IsUpload" type="text" /></div>
				</div>
				<div class="control-group">
					<label class="control-label" for="IsDelete">是否同步删除</label>
					<div class="controls"><input id="IsDelete" name="IsDelete" type="text" /></div>
				</div>
				<div class="control-group">
					<label class="control-label" for="LoadUpdateTime">下载更新时间</label>
					<div class="controls"><input id="LoadUpdateTime" name="LoadUpdateTime" type="text" /></div>
				</div>

				<div class="control-group">
					<div class="controls"><button class="btn" contenteditable="true" type="submit">添加</button></div>
				</div>
			</form>
		</div>
	</div>
</div>
		