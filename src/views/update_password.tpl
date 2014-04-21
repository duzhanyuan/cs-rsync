<div class="container-fluid">
	<div class="row-fluid">
		<div class="span12">
			<form class="form-horizontal" method="POST" action="/UpdatePassword">
				<div class="control-group">
					<label class="control-label" for="Password">旧密码</label>
					<div class="controls"><input id="Password" name="Password" type="password" /></div>
				</div>
				<div class="control-group">
					<label class="control-label" for="NewPassword">新密码</label>
					<div class="controls"><input id="NewPassword" name="NewPassword" type="password" /></div>
				</div>
				<div class="control-group">
					<label class="control-label" for="NewPasswordConfirm">新密码确认</label>
					<div class="controls"><input id="NewPasswordConfirm" name="NewPasswordConfirm" type="password" /></div>
				</div>
				
				<div class="control-group">
					<div class="controls"><button class="btn" contenteditable="true" type="submit">登录</button></div>
				</div>
			</form>
		</div>
	</div>
</div>