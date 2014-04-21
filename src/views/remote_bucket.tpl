<div class="container-fluid">
	<div class="row-fluid">
		<div class="hero-unit">
			<table class="table">
				<thead>
					<tr>
						<th>名称</th>
						<th>所在数据中心</th>
						<th>创建时间</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{{range .BucketList}}
					<tr>
						<td>{{.Name}}</td>
						<td>{{.Location}}</td>
						<td>{{.CreationDate}}</td>
						<td><button class="btn" type="button"><a href="/RemoteBucket?Operate=Delete&BucketName={{.Name}}" >删除</a></button></td>
					</tr>
					{{end}}
				</tbody>
			</table>
		</div>
		<div class="span12">
			<form class="form-horizontal" method="POST" action="/RemoteBucket?Operate=Add">
				<div class="control-group">
					<label class="control-label" contenteditable="true" for="BucketName">Bucket名称</label>
					<div class="controls"><input id="BucketName" name="BucketName" type="text" /></div>
				</div>
				
				<div class="control-group">
					<div class="controls"><button class="btn" contenteditable="true" type="submit">添加</button></div>
				</div>
			</form>
		</div>
	</div>
</div>