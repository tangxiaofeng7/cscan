<template>
  <div class="poc-page">
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- Nuclei默认模板 -->
      <el-tab-pane label="默认模板" name="nucleiTemplates">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>Nuclei 默认模板库</span>
              <span style="color: #909399; font-size: 13px; margin-left: 10px">
                共 {{ templateStats.total || 0 }} 个模板
              </span>
              <el-dropdown style="margin-left: auto" @command="handleSyncCommand">
                <el-button type="primary" size="small" :loading="syncLoading">
                  <el-icon><Refresh /></el-icon>同步模板<el-icon class="el-icon--right"><arrow-down /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="sync">增量同步</el-dropdown-item>
                    <el-dropdown-item command="force">强制重新同步</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
          <p class="tip-text">
            Nuclei 模板已同步到数据库，扫描时将根据任务配置的严重级别从数据库加载模板。程序启动时自动从 ~/nuclei-templates 同步。
          </p>
          <!-- 筛选条件 -->
          <el-form :inline="true" class="filter-form">
            <el-form-item label="分类">
              <el-select v-model="templateFilter.category" placeholder="全部分类" clearable style="width: 150px" @change="loadNucleiTemplates">
                <el-option v-for="cat in templateCategories" :key="cat" :label="cat" :value="cat" />
              </el-select>
            </el-form-item>
            <el-form-item label="级别">
              <el-select v-model="templateFilter.severity" placeholder="全部级别" clearable style="width: 120px" @change="loadNucleiTemplates">
                <el-option label="Critical" value="critical" />
                <el-option label="High" value="high" />
                <el-option label="Medium" value="medium" />
                <el-option label="Low" value="low" />
                <el-option label="Info" value="info" />
              </el-select>
            </el-form-item>
            <el-form-item label="标签">
              <el-input v-model="templateFilter.tag" placeholder="输入标签" clearable style="width: 150px" @keyup.enter="loadNucleiTemplates" />
            </el-form-item>
            <el-form-item label="搜索">
              <el-input v-model="templateFilter.keyword" placeholder="名称/ID/描述" clearable style="width: 180px" @keyup.enter="loadNucleiTemplates" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="loadNucleiTemplates">搜索</el-button>
            </el-form-item>
          </el-form>
          <!-- 统计信息和批量操作 -->
          <div class="stats-bar" v-if="templateStats.total">
            <el-tag type="danger" size="small">Critical: {{ templateStats.critical || 0 }}</el-tag>
            <el-tag type="warning" size="small">High: {{ templateStats.high || 0 }}</el-tag>
            <el-tag size="small">Medium: {{ templateStats.medium || 0 }}</el-tag>
            <el-tag type="info" size="small">Low: {{ templateStats.low || 0 }}</el-tag>
            <el-tag type="success" size="small">Info: {{ templateStats.info || 0 }}</el-tag>
            <el-button 
              v-if="selectedTemplates.length > 0" 
              type="success" 
              size="small" 
              style="margin-left: 20px"
              @click="showTemplateBatchValidateDialog"
            >
              批量验证 ({{ selectedTemplates.length }})
            </el-button>
          </div>
          <!-- 模板列表 -->
          <el-table 
            :data="nucleiTemplates" 
            stripe 
            v-loading="nucleiTemplateLoading" 
            max-height="500"
            @selection-change="handleTemplateSelectionChange"
          >
            <el-table-column type="selection" width="45" />
            <el-table-column prop="id" label="模板ID" width="200" show-overflow-tooltip />
            <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
            <el-table-column prop="severity" label="级别" width="90">
              <template #default="{ row }">
                <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="category" label="分类" width="100" />
            <el-table-column prop="tags" label="标签" min-width="180">
              <template #default="{ row }">
                <el-tag v-for="tag in (row.tags || [])" :key="tag" size="small" style="margin-right: 3px">
                  {{ tag }}
                </el-tag>
                <span v-if="row.tags && row.tags.length > 4" style="color: #909399">+{{ row.tags.length - 4 }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="author" label="作者" width="100" show-overflow-tooltip />
            <el-table-column label="操作" width="120" fixed="right">
              <template #default="{ row }">
                <el-button type="success" link size="small" @click="showTemplateValidateDialog(row)">验证</el-button>
                <el-button type="primary" link size="small" @click="showTemplateContent(row)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="templatePagination.page"
            v-model:page-size="templatePagination.pageSize"
            :total="templatePagination.total"
            :page-sizes="[50, 100, 200]"
            layout="total, sizes, prev, pager, next"
            class="pagination"
            @size-change="loadNucleiTemplates"
            @current-change="loadNucleiTemplates"
          />
        </el-card>
      </el-tab-pane>

      <!-- 标签映射 -->
      <el-tab-pane label="标签映射" name="tagMapping">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>应用标签映射配置</span>
              <span style="color: #909399; font-size: 13px; margin-left: 10px">
                共 {{ tagMappings.length || 0 }} 条映射
              </span>
              <el-button type="primary" size="small" style="margin-left: auto" @click="showTagMappingForm()">
                <el-icon><Plus /></el-icon>新增映射
              </el-button>
            </div>
          </template>
          <p class="tip-text">
            配置 Wappalyzer 识别的应用名称与 Nuclei 标签的映射关系，扫描时会根据识别到的应用自动选择对应的 POC 进行检测。
          </p>
          <el-table :data="tagMappings" stripe v-loading="tagMappingLoading" max-height="500">
            <el-table-column prop="appName" label="应用名称" width="180" />
            <el-table-column prop="nucleiTags" label="POC标签（Tag）" min-width="250">
              <template #default="{ row }">
                <el-tag v-for="tag in row.nucleiTags" :key="tag" size="small" style="margin-right: 5px">
                  {{ tag }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="150" />
            <el-table-column prop="enabled" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                  {{ row.enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="showTagMappingForm(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteTagMapping(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 自定义POC -->
      <el-tab-pane label="自定义POC" name="customPoc">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>自定义 Nuclei POC</span>
              <span style="color: #909399; font-size: 13px; margin-left: 10px">
                共 {{ pocPagination.total || 0 }} 个POC
              </span>
              <div style="margin-left: auto">
                <el-button type="danger" size="small" @click="handleClearAllPocs" :loading="clearPocLoading" style="margin-right: 10px">
                  <el-icon><Delete /></el-icon>清空
                </el-button>
                <el-button type="warning" size="small" @click="handleExportPocs" :loading="exportPocLoading" style="margin-right: 10px">
                  <el-icon><Download /></el-icon>导出POC
                </el-button>
                <el-button type="success" size="small" @click="showImportPocDialog" style="margin-right: 10px">
                  <el-icon><Upload /></el-icon>导入POC
                </el-button>
                <el-button type="primary" size="small" @click="showCustomPocForm()">
                  <el-icon><Plus /></el-icon>新增POC
                </el-button>
              </div>
            </div>
          </template>
          <!-- 筛选条件 -->
          <el-form :inline="true" class="filter-form">
            <el-form-item label="名称">
              <el-input v-model="customPocFilter.name" placeholder="POC名称" clearable style="width: 150px" @keyup.enter="loadCustomPocs" />
            </el-form-item>
            <el-form-item label="模板ID">
              <el-input v-model="customPocFilter.templateId" placeholder="模板ID" clearable style="width: 150px" @keyup.enter="loadCustomPocs" />
            </el-form-item>
            <el-form-item label="级别">
              <el-select v-model="customPocFilter.severity" placeholder="全部级别" clearable style="width: 120px" @change="loadCustomPocs">
                <el-option label="Critical" value="critical" />
                <el-option label="High" value="high" />
                <el-option label="Medium" value="medium" />
                <el-option label="Low" value="low" />
                <el-option label="Info" value="info" />
              </el-select>
            </el-form-item>
            <el-form-item label="标签">
              <el-input v-model="customPocFilter.tag" placeholder="输入标签" clearable style="width: 120px" @keyup.enter="loadCustomPocs" />
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="customPocFilter.enabled" placeholder="全部状态" clearable style="width: 100px" @change="loadCustomPocs">
                <el-option label="启用" :value="true" />
                <el-option label="禁用" :value="false" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="loadCustomPocs">搜索</el-button>
              <el-button @click="resetCustomPocFilter">重置</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="customPocs" stripe v-loading="customPocLoading" max-height="500">
            <el-table-column prop="name" label="名称" width="250" />
            <el-table-column prop="templateId" label="模板ID" width="250" />
            <el-table-column prop="severity" label="严重级别" width="100">
              <template #default="{ row }">
                <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="tags" label="标签" min-width="200">
              <template #default="{ row }">
                <el-tag v-for="tag in row.tags" :key="tag" size="small" style="margin-right: 5px">
                  {{ tag }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="enabled" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                  {{ row.enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button type="success" link size="small" @click="showPocValidateDialog(row)">验证</el-button>
                <el-button type="primary" link size="small" @click="showCustomPocForm(row)">编辑</el-button>
                <el-button type="danger" link size="small" @click="handleDeleteCustomPoc(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            v-model:current-page="pocPagination.page"
            v-model:page-size="pocPagination.pageSize"
            :total="pocPagination.total"
            :page-sizes="[20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            class="pagination"
            @size-change="loadCustomPocs"
            @current-change="loadCustomPocs"
          />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 标签映射编辑对话框 -->
    <el-dialog v-model="tagMappingDialogVisible" :title="tagMappingForm.id ? '编辑映射' : '新增映射'" width="500px">
      <el-form ref="tagMappingFormRef" :model="tagMappingForm" :rules="tagMappingRules" label-width="100px">
        <el-form-item label="应用名称" prop="appName">
          <el-input v-model="tagMappingForm.appName" placeholder="Wappalyzer识别的应用名称，如: WordPress" />
        </el-form-item>
        <el-form-item label="Nuclei标签" prop="nucleiTagsInput">
          <el-input 
            v-model="tagMappingForm.nucleiTagsInput" 
            placeholder="输入Nuclei标签，多个用逗号分隔，如: wordpress,wp-plugin,cve"
            style="width: 100%"
          />
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            常用标签: wordpress, apache, nginx, php, java, cve, rce, sqli, xss, lfi
          </div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="tagMappingForm.description" placeholder="可选描述" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="tagMappingForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="tagMappingDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveTagMapping">保存</el-button>
      </template>
    </el-dialog>

    <!-- 自定义POC编辑对话框 -->
    <el-dialog v-model="customPocDialogVisible" :title="customPocForm.id ? '编辑POC' : '新增POC'" width="900px">
      <el-form ref="customPocFormRef" :model="customPocForm" :rules="customPocRules" label-width="100px">
        <el-form-item label="YAML内容" prop="content">
          <div style="width: 100%">
            <div style="margin-bottom: 8px; color: #909399; font-size: 12px">
              粘贴或编辑 Nuclei YAML 模板，下方字段将自动从 YAML 中解析
            </div>
            <div class="yaml-editor-wrapper">
              <el-input
                v-model="customPocForm.content"
                type="textarea"
                :rows="18"
                placeholder="Nuclei YAML模板内容"
                @input="parseYamlContent"
              />
            </div>
          </div>
        </el-form-item>
        <el-divider content-position="left">解析结果（自动从YAML提取，可手动修改）</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="模板ID" prop="templateId">
              <el-input v-model="customPocForm.templateId" placeholder="从YAML的id字段解析" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="名称" prop="name">
              <el-input v-model="customPocForm.name" placeholder="从YAML的info.name解析" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="严重级别" prop="severity">
              <el-select v-model="customPocForm.severity" style="width: 100%">
                <el-option label="Critical" value="critical" />
                <el-option label="High" value="high" />
                <el-option label="Medium" value="medium" />
                <el-option label="Low" value="low" />
                <el-option label="Info" value="info" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="作者">
              <el-input v-model="customPocForm.author" placeholder="从YAML的info.author解析" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="标签">
          <el-input 
            v-model="customPocForm.tagsInput" 
            placeholder="从YAML的info.tags解析，多个用逗号分隔"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="customPocForm.description" type="textarea" :rows="2" placeholder="从YAML的info.description解析" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="customPocForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="customPocDialogVisible = false">取消</el-button>
        <el-button @click="parseYamlContent">重新解析YAML</el-button>
        <el-button type="primary" @click="handleSaveCustomPoc">保存</el-button>
      </template>
    </el-dialog>

    <!-- 导入POC对话框 -->
    <el-dialog v-model="importPocDialogVisible" title="导入POC" width="900px">
      <el-form label-width="100px">
        <el-form-item label="POC格式">
          <el-radio-group v-model="importPocFormat">
            <el-radio-button value="nuclei">Nuclei</el-radio-button>
            <el-radio-button value="xray">XRAY (自动转换)</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="导入方式">
          <el-radio-group v-model="importPocType">
            <el-radio-button value="text">文本粘贴</el-radio-button>
            <el-radio-button value="file">文件上传</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="importPocType === 'text'" label="POC内容">
          <div style="width: 100%">
            <div style="margin-bottom: 8px; color: #909399; font-size: 12px">
              {{ importPocFormat === 'xray' ? '粘贴 XRAY YAML POC 内容，将自动转换为 Nuclei 格式' : '粘贴 Nuclei YAML 模板内容' }}，支持一次导入多个POC（用 --- 分隔）
            </div>
            <div class="yaml-editor-wrapper">
              <el-input
                v-model="importPocContent"
                type="textarea"
                :rows="18"
                :placeholder="importPocFormat === 'xray' ? '粘贴 XRAY YAML POC 内容...' : '粘贴 Nuclei YAML 模板内容...'"
                @input="parseImportContent"
              />
            </div>
          </div>
        </el-form-item>
        <el-form-item v-else label="上传文件">
          <div style="width: 100%">
            <el-upload
              ref="importPocUploadRef"
              :auto-upload="false"
              :limit="500"
              accept=".yaml,.yml"
              :on-change="handleImportFileChange"
              :on-remove="handleImportFileRemove"
              :before-upload="() => false"
              multiple
              drag
              :show-file-list="false"
            >
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">拖拽文件到此处，或 <em>点击上传</em></div>
              <template #tip>
                <div class="el-upload__tip">
                  支持 .yaml / .yml 文件，可批量选择多个文件
                  <span v-if="importPocFormat === 'xray'" style="color: #e6a23c">（XRAY格式将自动转换为Nuclei格式）</span>
                </div>
              </template>
            </el-upload>
            <div v-if="uploadedFileCount > 0" style="margin-top: 10px; color: #67c23a; font-size: 13px">
              <el-icon><UploadFilled /></el-icon> 已上传 {{ uploadedFileCount }} 个文件
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <!-- 解析预览 -->
      <div v-if="importPocPreviews.length > 0" class="import-preview">
        <div class="preview-header">
          <span>解析预览 ({{ importPocPreviews.length }} 个POC)</span>
          <el-tag v-if="importPocFormat === 'xray'" type="warning" size="small" style="margin-left: 10px">已转换为Nuclei格式</el-tag>
          <el-checkbox v-model="importPocEnabled" style="margin-left: 15px">导入后启用</el-checkbox>
        </div>
        <el-table :data="importPocPreviews" max-height="300" size="small">
          <el-table-column prop="templateId" label="模板ID" width="180" show-overflow-tooltip />
          <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
          <el-table-column prop="severity" label="级别" width="90">
            <template #default="{ row }">
              <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="author" label="作者" width="100" show-overflow-tooltip />
          <el-table-column prop="tags" label="标签" min-width="150">
            <template #default="{ row }">
              <el-tag v-for="tag in (row.tags || []).slice(0, 3)" :key="tag" size="small" style="margin-right: 3px">
                {{ tag }}
              </el-tag>
              <span v-if="row.tags && row.tags.length > 3" style="color: #909399">+{{ row.tags.length - 3 }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="{ row, $index }">
              <el-button type="primary" link size="small" @click="previewConvertedPoc(row)">预览</el-button>
              <el-button type="danger" link size="small" @click="removeImportPreview($index)">移除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <template #footer>
        <el-button @click="importPocDialogVisible = false">取消</el-button>
        <el-button @click="clearImportContent">清空</el-button>
        <el-button type="primary" @click="handleImportPocs" :loading="importPocLoading" :disabled="importPocPreviews.length === 0">
          导入 ({{ importPocPreviews.length }})
        </el-button>
      </template>
    </el-dialog>

    <!-- 预览转换后的POC对话框 -->
    <el-dialog v-model="convertedPocPreviewVisible" title="转换后的POC预览" width="800px">
      <el-input
        v-model="convertedPocPreviewContent"
        type="textarea"
        :rows="25"
        readonly
        style="font-family: 'Consolas', 'Monaco', monospace; font-size: 13px"
      />
      <template #footer>
        <el-button @click="convertedPocPreviewVisible = false">关闭</el-button>
        <el-button type="primary" @click="copyConvertedPoc">复制内容</el-button>
      </template>
    </el-dialog>

    <!-- 查看模板内容对话框 -->
    <el-dialog v-model="templateContentDialogVisible" :title="currentTemplate.name || '模板内容'" width="900px">
      <el-descriptions :column="2" border size="small" style="margin-bottom: 15px">
        <el-descriptions-item label="模板ID">{{ currentTemplate.id }}</el-descriptions-item>
        <el-descriptions-item label="严重级别">
          <el-tag :type="getSeverityType(currentTemplate.severity)" size="small">{{ currentTemplate.severity }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="分类">{{ currentTemplate.category }}</el-descriptions-item>
        <el-descriptions-item label="作者">{{ currentTemplate.author }}</el-descriptions-item>
        <el-descriptions-item label="标签" :span="2">
          <el-tag v-for="tag in (currentTemplate.tags || [])" :key="tag" size="small" style="margin-right: 5px">{{ tag }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ currentTemplate.description || '-' }}</el-descriptions-item>
      </el-descriptions>
      <div class="template-content-wrapper">
        <el-input
          v-model="currentTemplate.content"
          type="textarea"
          :rows="20"
          readonly
          style="font-family: 'Consolas', 'Monaco', monospace; font-size: 13px"
        />
      </div>
      <template #footer>
        <el-button @click="templateContentDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="copyTemplateContent">复制内容</el-button>
      </template>
    </el-dialog>

    <!-- POC验证对话框 -->
    <el-dialog v-model="pocValidateDialogVisible" title="验证POC" width="700px" @close="handleValidateDialogClose">
      <el-form label-width="80px">
        <el-form-item label="POC名称">
          <el-input :value="validatePoc.name" disabled />
        </el-form-item>
        <el-form-item label="模板ID">
          <el-input :value="validatePoc.templateId" disabled />
        </el-form-item>
        <el-form-item label="目标URL">
          <el-input v-model="pocValidateUrl" placeholder="请输入目标URL，如 https://example.com" />
        </el-form-item>
      </el-form>
      
      <!-- 执行日志区域 -->
      <div v-if="pocValidateLoading || pocValidateLogs.length > 0" class="validate-logs">
        <div class="logs-header">
          <span>执行日志</span>
          <el-tag v-if="pocValidateLoading" type="warning" size="small">执行中...</el-tag>
          <el-tag v-else-if="pocValidateResult && pocValidateResult.matched" type="success" size="small">发现漏洞</el-tag>
          <el-tag v-else-if="pocValidateResult" type="info" size="small">完成</el-tag>
        </div>
        <div class="logs-content" ref="logsContainerRef">
          <div v-for="(log, index) in pocValidateLogs" :key="index" :class="['log-line', 'log-' + log.level.toLowerCase()]">
            <span class="log-time">{{ log.timestamp }}</span>
            <span class="log-level">[{{ log.level }}]</span>
            <span class="log-msg">{{ log.message }}</span>
          </div>
        </div>
      </div>
      
      <!-- 验证结果区域 -->
      <div v-if="pocValidateResult && !pocValidateLoading" class="validate-result">
        <div class="result-header">
          <el-tag :type="pocValidateResult.matched ? 'danger' : 'info'" size="large">
            {{ pocValidateResult.matched ? '✓ 发现漏洞' : '✗ 未发现漏洞' }}
          </el-tag>
          <el-tag :type="getSeverityType(pocValidateResult.severity)" size="small" style="margin-left: 10px">
            {{ pocValidateResult.severity }}
          </el-tag>
        </div>
        <pre class="result-details">{{ pocValidateResult.details }}</pre>
      </div>
      <template #footer>
        <el-button @click="pocValidateDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleValidatePoc" :loading="pocValidateLoading" :disabled="!pocValidateUrl">验证</el-button>
      </template>
    </el-dialog>

    <!-- 默认模板批量验证对话框 -->
    <el-dialog v-model="templateBatchValidateDialogVisible" title="批量验证POC" width="900px" @close="handleBatchValidateDialogClose">
      <el-form label-width="100px">
        <el-form-item label="已选模板">
          <div class="selected-templates">
            <el-tag v-for="tpl in selectedTemplates.slice(0, 10)" :key="tpl.id" size="small" style="margin-right: 5px; margin-bottom: 5px">
              {{ tpl.name || tpl.id }}
            </el-tag>
            <span v-if="selectedTemplates.length > 10" style="color: #909399">+{{ selectedTemplates.length - 10 }} 个</span>
          </div>
        </el-form-item>
        <el-form-item label="目标URL">
          <div style="width: 100%">
            <div style="margin-bottom: 8px; display: flex; align-items: center; gap: 10px;">
              <el-radio-group v-model="batchTargetInputType" size="small">
                <el-radio-button value="text">文本输入</el-radio-button>
                <el-radio-button value="file">文件上传</el-radio-button>
              </el-radio-group>
              <span style="color: #909399; font-size: 12px;">
                {{ batchTargetInputType === 'text' ? '每行一个URL' : '支持 .txt 文件，每行一个URL' }}
              </span>
            </div>
            <el-input 
              v-if="batchTargetInputType === 'text'"
              v-model="templateBatchValidateUrls" 
              type="textarea" 
              :rows="5" 
              placeholder="请输入目标URL，每行一个，如：&#10;https://example1.com&#10;https://example2.com&#10;https://example3.com"
            />
            <el-upload
              v-else
              ref="batchUrlUploadRef"
              :auto-upload="false"
              :limit="1"
              accept=".txt"
              :on-change="handleBatchUrlFileChange"
              :on-remove="handleBatchUrlFileRemove"
              drag
            >
              <el-icon class="el-icon--upload"><upload-filled /></el-icon>
              <div class="el-upload__text">拖拽文件到此处，或 <em>点击上传</em></div>
              <template #tip>
                <div class="el-upload__tip">仅支持 .txt 文件，每行一个URL</div>
              </template>
            </el-upload>
            <div v-if="batchTargetUrls.length > 0" style="margin-top: 8px; color: #67c23a; font-size: 12px;">
              已解析 {{ batchTargetUrls.length }} 个目标URL
            </div>
          </div>
        </el-form-item>
      </el-form>
      
      <!-- 批量验证进度 -->
      <div v-if="templateBatchValidateLoading || templateBatchValidateResults.length > 0" class="batch-validate-progress">
        <div class="progress-header">
          <span>验证进度: {{ templateBatchValidateProgress.completed }}/{{ templateBatchValidateProgress.total }}</span>
          <el-progress 
            :percentage="templateBatchValidateProgress.total > 0 ? Math.round(templateBatchValidateProgress.completed / templateBatchValidateProgress.total * 100) : 0" 
            :status="templateBatchValidateLoading ? '' : 'success'"
            style="width: 200px; margin-left: 15px"
          />
        </div>
        
        <!-- 执行日志 -->
        <div class="logs-content" ref="batchLogsContainerRef" style="max-height: 150px;">
          <div v-for="(log, index) in templateBatchValidateLogs" :key="index" :class="['log-line', 'log-' + log.level.toLowerCase()]">
            <span class="log-time">{{ log.timestamp }}</span>
            <span class="log-level">[{{ log.level }}]</span>
            <span class="log-msg">{{ log.message }}</span>
          </div>
        </div>
      </div>
      
      <!-- 批量验证结果 -->
      <div v-if="templateBatchValidateResults.length > 0" class="batch-validate-results">
        <div class="results-header">
          <span>验证结果</span>
          <el-tag type="danger" size="small" style="margin-left: 10px">
            发现漏洞: {{ templateBatchValidateResults.filter(r => r.matched).length }}
          </el-tag>
          <el-tag type="info" size="small" style="margin-left: 5px">
            未匹配: {{ templateBatchValidateResults.filter(r => !r.matched).length }}
          </el-tag>
          <el-dropdown style="margin-left: auto" @command="handleExportResults">
            <el-button type="success" size="small">
              导出结果<el-icon class="el-icon--right"><arrow-down /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="all">导出全部</el-dropdown-item>
                <el-dropdown-item command="matched">仅导出匹配</el-dropdown-item>
                <el-dropdown-item command="csv">导出CSV</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        <el-table :data="templateBatchValidateResults" max-height="250" size="small">
          <el-table-column prop="pocName" label="模板名称" min-width="150" show-overflow-tooltip />
          <el-table-column prop="severity" label="级别" width="80">
            <template #default="{ row }">
              <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="matched" label="结果" width="80">
            <template #default="{ row }">
              <el-tag :type="row.matched ? 'danger' : 'info'" size="small">
                {{ row.matched ? '匹配' : '未匹配' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="matchedUrl" label="匹配URL" min-width="200" show-overflow-tooltip />
        </el-table>
      </div>
      
      <template #footer>
        <el-button @click="templateBatchValidateDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="handleTemplateBatchValidate" :loading="templateBatchValidateLoading" :disabled="batchTargetUrls.length === 0">
          开始验证
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, ArrowDown, UploadFilled, Upload, Download, Delete } from '@element-plus/icons-vue'
import { getTagMappingList, saveTagMapping, deleteTagMapping, getCustomPocList, saveCustomPoc, batchImportCustomPoc, deleteCustomPoc, clearAllCustomPoc, getNucleiTemplateList, getNucleiTemplateCategories, syncNucleiTemplates, getNucleiTemplateDetail, validatePoc as validatePocApi, getPocValidationResult } from '@/api/poc'
import jsYaml from 'js-yaml'
import JSZip from 'jszip'
import { saveAs } from 'file-saver'

const activeTab = ref('nucleiTemplates')

// Nuclei默认模板
const nucleiTemplates = ref([])
const nucleiTemplateLoading = ref(false)
const templateCategories = ref([])
const templateTags = ref([])
const selectedTemplates = ref([])
const templateStats = ref({})
const templateFilter = reactive({
  category: '',
  severity: '',
  tag: '',
  keyword: ''
})
const templatePagination = reactive({
  page: 1,
  pageSize: 50,
  total: 0
})
const syncLoading = ref(false)
const templateContentDialogVisible = ref(false)
const currentTemplate = ref({})

// 标签映射
const tagMappings = ref([])
const tagMappingLoading = ref(false)
const tagMappingDialogVisible = ref(false)
const tagMappingFormRef = ref()
const tagMappingForm = reactive({
  id: '',
  appName: '',
  nucleiTags: [],
  nucleiTagsInput: '', // 用户输入的逗号分隔标签
  description: '',
  enabled: true
})
const tagMappingRules = {
  appName: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
  nucleiTagsInput: [{ required: true, message: '请输入Nuclei标签', trigger: 'blur' }]
}

// 自定义POC
const customPocs = ref([])
const customPocLoading = ref(false)
const customPocDialogVisible = ref(false)

// 自定义POC筛选条件
const customPocFilter = reactive({
  name: '',
  templateId: '',
  severity: '',
  tag: '',
  enabled: null
})

// 导入POC
const importPocDialogVisible = ref(false)
const importPocType = ref('text')
const importPocFormat = ref('nuclei') // nuclei 或 xray
const importPocContent = ref('')
const importPocPreviews = ref([])
const importPocEnabled = ref(true)
const importPocLoading = ref(false)
const uploadedFileCount = ref(0)
const importPocUploadRef = ref(null)
const convertedPocPreviewVisible = ref(false)
const convertedPocPreviewContent = ref('')
const exportPocLoading = ref(false)
const clearPocLoading = ref(false)
const customPocFormRef = ref()
const customPocForm = reactive({
  id: '',
  name: '',
  templateId: '',
  severity: 'medium',
  tags: [],
  tagsInput: '', // 用户输入的逗号分隔标签
  author: '',
  description: '',
  content: '',
  enabled: true
})
const customPocRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  templateId: [{ required: true, message: '请输入模板ID', trigger: 'blur' }],
  severity: [{ required: true, message: '请选择严重级别', trigger: 'change' }],
  content: [{ required: true, message: '请输入YAML内容', trigger: 'blur' }]
}
const pocPagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// POC验证
const pocValidateDialogVisible = ref(false)
const validatePoc = ref({})
const pocValidateUrl = ref('')
const pocValidateResult = ref(null)
const pocValidateLoading = ref(false)
const pocValidateLogs = ref([])
const logsContainerRef = ref(null)
let logEventSource = null
let currentTaskId = null

// 默认模板批量验证
const templateBatchValidateDialogVisible = ref(false)
const templateBatchValidateUrls = ref('') // 多行URL文本
const batchTargetInputType = ref('text') // 输入类型: text 或 file
const batchUrlUploadRef = ref(null)
const templateBatchValidateLoading = ref(false)
const templateBatchValidateLogs = ref([])
const templateBatchValidateResults = ref([])
const templateBatchValidateProgress = reactive({ total: 0, completed: 0 })
const batchLogsContainerRef = ref(null)
const currentBatchTaskIds = ref([]) // 当前批次的任务ID列表
let batchLogEventSource = null
let batchPollTimer = null
let currentBatchId = null

// 计算属性：解析多行URL为数组（自动去重）
const batchTargetUrls = computed(() => {
  if (!templateBatchValidateUrls.value) return []
  const urls = templateBatchValidateUrls.value
    .split('\n')
    .map(url => url.trim())
    .filter(url => url && (url.startsWith('http://') || url.startsWith('https://')))
  // 使用Set去重
  return [...new Set(urls)]
})

onMounted(() => {
  // 只加载当前标签页需要的数据
  loadNucleiTemplateCategories()
  loadNucleiTemplates()
})

function handleTabChange(tab) {
  if (tab === 'nucleiTemplates' && nucleiTemplates.value.length === 0) {
    loadNucleiTemplateCategories()
    loadNucleiTemplates()
  } else if (tab === 'tagMapping' && tagMappings.value.length === 0) {
    loadTagMappings()
  } else if (tab === 'customPoc' && customPocs.value.length === 0) {
    loadCustomPocs()
  }
}

async function loadNucleiTemplateCategories() {
  try {
    const res = await getNucleiTemplateCategories()
    if (res.code === 0) {
      templateCategories.value = res.categories || []
      templateTags.value = res.tags || []
      templateStats.value = res.stats || {}
    }
  } catch (e) {
    console.error('Failed to load template categories:', e)
  }
}

async function loadNucleiTemplates() {
  nucleiTemplateLoading.value = true
  try {
    const res = await getNucleiTemplateList({
      category: templateFilter.category,
      severity: templateFilter.severity,
      tag: templateFilter.tag,
      keyword: templateFilter.keyword,
      page: templatePagination.page,
      pageSize: templatePagination.pageSize
    })
    if (res.code === 0) {
      nucleiTemplates.value = res.list || []
      templatePagination.total = res.total
    } else {
      ElMessage.error(res.msg || '加载模板失败')
    }
  } finally {
    nucleiTemplateLoading.value = false
  }
}

async function handleSyncCommand(command) {
  if (command === 'force') {
    try {
      await ElMessageBox.confirm('强制同步将删除所有现有模板并重新导入，确定继续吗？', '提示', { type: 'warning' })
    } catch {
      return
    }
  }
  
  syncLoading.value = true
  try {
    const res = await syncNucleiTemplates({ force: command === 'force' })
    if (res.code === 0) {
      ElMessage.success(res.msg)
      // 延迟刷新数据
      setTimeout(() => {
        loadNucleiTemplateCategories()
        loadNucleiTemplates()
      }, 3000)
    } else {
      ElMessage.error(res.msg)
    }
  } finally {
    syncLoading.value = false
  }
}

async function showTemplateContent(row) {
  // 需要从API获取完整内容
  const res = await getNucleiTemplateDetail({ templateId: row.id })
  if (res.code === 0 && res.data) {
    currentTemplate.value = res.data
    // 如果内容为空，提示用户强制同步
    if (!res.data.content) {
      currentTemplate.value.content = '# YAML内容为空\n# 请点击"同步模板" -> "强制重新同步"来更新模板内容'
    }
  } else {
    currentTemplate.value = { ...row, content: '加载失败，请重试' }
  }
  templateContentDialogVisible.value = true
}

function copyTemplateContent() {
  if (currentTemplate.value.content) {
    navigator.clipboard.writeText(currentTemplate.value.content)
    ElMessage.success('已复制到剪贴板')
  }
}

async function loadTagMappings() {
  tagMappingLoading.value = true
  try {
    const res = await getTagMappingList()
    if (res.code === 0) {
      tagMappings.value = res.list || []
    }
  } finally {
    tagMappingLoading.value = false
  }
}

async function loadCustomPocs() {
  customPocLoading.value = true
  try {
    const params = {
      page: pocPagination.page,
      pageSize: pocPagination.pageSize
    }
    // 添加筛选条件
    if (customPocFilter.name) {
      params.name = customPocFilter.name
    }
    if (customPocFilter.templateId) {
      params.templateId = customPocFilter.templateId
    }
    if (customPocFilter.severity) {
      params.severity = customPocFilter.severity
    }
    if (customPocFilter.tag) {
      params.tag = customPocFilter.tag
    }
    if (customPocFilter.enabled !== null && customPocFilter.enabled !== '') {
      params.enabled = customPocFilter.enabled
    }
    
    const res = await getCustomPocList(params)
    if (res.code === 0) {
      customPocs.value = res.list || []
      pocPagination.total = res.total
    }
  } finally {
    customPocLoading.value = false
  }
}

// 重置自定义POC筛选条件
function resetCustomPocFilter() {
  customPocFilter.name = ''
  customPocFilter.templateId = ''
  customPocFilter.severity = ''
  customPocFilter.tag = ''
  customPocFilter.enabled = null
  pocPagination.page = 1
  loadCustomPocs()
}

function showTagMappingForm(row = null) {
  if (row) {
    Object.assign(tagMappingForm, {
      id: row.id,
      appName: row.appName,
      nucleiTags: row.nucleiTags || [],
      nucleiTagsInput: (row.nucleiTags || []).join(', '), // 转换为逗号分隔字符串
      description: row.description,
      enabled: row.enabled
    })
  } else {
    Object.assign(tagMappingForm, {
      id: '',
      appName: '',
      nucleiTags: [],
      nucleiTagsInput: '',
      description: '',
      enabled: true
    })
  }
  tagMappingDialogVisible.value = true
}

async function handleSaveTagMapping() {
  await tagMappingFormRef.value.validate()
  // 将逗号分隔的字符串转换为数组
  const tagsArray = tagMappingForm.nucleiTagsInput
    .split(/[,，]/) // 支持中英文逗号
    .map(tag => tag.trim())
    .filter(tag => tag !== '')
  
  const submitData = {
    id: tagMappingForm.id,
    appName: tagMappingForm.appName,
    nucleiTags: tagsArray,
    description: tagMappingForm.description,
    enabled: tagMappingForm.enabled
  }
  
  const res = await saveTagMapping(submitData)
  if (res.code === 0) {
    ElMessage.success('保存成功')
    tagMappingDialogVisible.value = false
    loadTagMappings()
  } else {
    ElMessage.error(res.msg)
  }
}

async function handleDeleteTagMapping(row) {
  await ElMessageBox.confirm('确定删除该映射吗？', '提示', { type: 'warning' })
  const res = await deleteTagMapping({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    loadTagMappings()
  }
}

function showCustomPocForm(row = null) {
  if (row) {
    Object.assign(customPocForm, {
      id: row.id,
      name: row.name,
      templateId: row.templateId,
      severity: row.severity,
      tags: row.tags || [],
      tagsInput: (row.tags || []).join(', '), // 转换为逗号分隔字符串
      author: row.author,
      description: row.description,
      content: row.content,
      enabled: row.enabled
    })
  } else {
    Object.assign(customPocForm, {
      id: '',
      name: '',
      templateId: '',
      severity: 'medium',
      tags: [],
      tagsInput: '',
      author: '',
      description: '',
      content: getNucleiTemplate(),
      enabled: true
    })
    // 新建时自动解析默认模板
    parseYamlContent()
  }
  customPocDialogVisible.value = true
}

async function handleSaveCustomPoc() {
  await customPocFormRef.value.validate()
  // 将逗号分隔的字符串转换为数组
  const tagsArray = customPocForm.tagsInput
    .split(/[,，]/) // 支持中英文逗号
    .map(tag => tag.trim())
    .filter(tag => tag !== '')
  
  const submitData = {
    id: customPocForm.id,
    name: customPocForm.name,
    templateId: customPocForm.templateId,
    severity: customPocForm.severity,
    tags: tagsArray,
    author: customPocForm.author,
    description: customPocForm.description,
    content: customPocForm.content,
    enabled: customPocForm.enabled
  }
  
  const res = await saveCustomPoc(submitData)
  if (res.code === 0) {
    ElMessage.success('保存成功')
    customPocDialogVisible.value = false
    loadCustomPocs()
  } else {
    ElMessage.error(res.msg)
  }
}

async function handleDeleteCustomPoc(row) {
  await ElMessageBox.confirm('确定删除该POC吗？', '提示', { type: 'warning' })
  const res = await deleteCustomPoc({ id: row.id })
  if (res.code === 0) {
    ElMessage.success('删除成功')
    loadCustomPocs()
  }
}

// ==================== 导出POC相关函数 ====================

// 导出所有自定义POC（每个POC一个文件，打包成ZIP）
async function handleExportPocs() {
  if (customPocs.value.length === 0) {
    ElMessage.warning('没有可导出的POC')
    return
  }
  
  exportPocLoading.value = true
  
  try {
    // 获取所有POC（可能需要分页获取全部）
    let allPocs = []
    
    // 如果当前页数据不是全部，需要获取全部数据
    if (pocPagination.total > customPocs.value.length) {
      const res = await getCustomPocList({ page: 1, pageSize: pocPagination.total })
      if (res.code === 0) {
        allPocs = res.list || []
      } else {
        allPocs = customPocs.value
      }
    } else {
      allPocs = customPocs.value
    }
    
    if (allPocs.length === 0) {
      ElMessage.warning('没有可导出的POC')
      return
    }
    
    // 创建ZIP文件
    const zip = new JSZip()
    
    // 每个POC创建一个单独的文件
    for (const poc of allPocs) {
      // 使用templateId作为文件名，清理非法字符
      const fileName = (poc.templateId || poc.name || 'poc')
        .replace(/[<>:"/\\|?*]/g, '-')
        .replace(/\s+/g, '-')
      zip.file(`${fileName}.yaml`, poc.content)
    }
    
    // 生成ZIP并下载
    const content = await zip.generateAsync({ type: 'blob' })
    const dateStr = new Date().toISOString().slice(0, 10)
    saveAs(content, `custom-pocs-${dateStr}.zip`)
    
    ElMessage.success(`成功导出 ${allPocs.length} 个POC`)
  } catch (e) {
    console.error('Export error:', e)
    ElMessage.error('导出失败')
  } finally {
    exportPocLoading.value = false
  }
}

// 清空所有自定义POC
async function handleClearAllPocs() {
  if (customPocs.value.length === 0 && pocPagination.total === 0) {
    ElMessage.warning('没有可清空的POC')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要清空所有自定义POC吗？共 ${pocPagination.total} 个POC将被删除，此操作不可恢复！`,
      '危险操作',
      {
        type: 'warning',
        confirmButtonText: '确定清空',
        cancelButtonText: '取消',
        confirmButtonClass: 'el-button--danger'
      }
    )
    
    clearPocLoading.value = true
    
    const res = await clearAllCustomPoc()
    if (res.code === 0) {
      ElMessage.success(`成功清空 ${res.deleted || pocPagination.total} 个POC`)
      loadCustomPocs()
    } else {
      ElMessage.error(res.msg || '清空失败')
    }
  } catch (e) {
    if (e !== 'cancel') {
      console.error('Clear error:', e)
      ElMessage.error('清空失败')
    }
  } finally {
    clearPocLoading.value = false
  }
}

// ==================== 导入POC相关函数 ====================

// 显示导入对话框
function showImportPocDialog() {
  importPocType.value = 'text'
  importPocFormat.value = 'nuclei'
  importPocContent.value = ''
  importPocPreviews.value = []
  importPocEnabled.value = true
  uploadedFileCount.value = 0
  importPocDialogVisible.value = true
}

// 解析导入的YAML内容
function parseImportContent() {
  if (!importPocContent.value.trim()) {
    importPocPreviews.value = []
    return
  }
  
  // 支持多个POC用 --- 分隔
  const yamlDocs = importPocContent.value.split(/\n---\s*\n/)
  const previews = []
  const seenTemplateIds = new Set()
  const seenContents = new Set()
  
  for (const doc of yamlDocs) {
    const trimmedDoc = doc.trim()
    if (!trimmedDoc) continue
    
    let parsed
    if (importPocFormat.value === 'xray') {
      parsed = parseXrayToNuclei(trimmedDoc)
    } else {
      parsed = parseYamlToPreview(trimmedDoc)
    }
    if (parsed) {
      // 检查是否已存在相同templateId或相同内容
      const contentHash = parsed.content.trim()
      if (!seenTemplateIds.has(parsed.templateId) && !seenContents.has(contentHash)) {
        seenTemplateIds.add(parsed.templateId)
        seenContents.add(contentHash)
        previews.push(parsed)
      }
    }
  }
  
  importPocPreviews.value = previews
}

// 解析单个YAML文档为预览对象
function parseYamlToPreview(content) {
  const result = {
    templateId: '',
    name: '',
    author: '',
    severity: 'medium',
    description: '',
    tags: [],
    content: content
  }
  
  // 解析 id 字段
  const idMatch = content.match(/^id:\s*(.+)$/m)
  if (idMatch) {
    result.templateId = idMatch[1].trim()
  }
  
  // 解析 info 块中的字段
  const infoMatch = content.match(/info:\s*\n((?:\s+.+\n?)+)/m)
  if (infoMatch) {
    const infoBlock = infoMatch[1]
    
    // name
    const nameMatch = infoBlock.match(/^\s+name:\s*(.+)$/m)
    if (nameMatch) {
      result.name = nameMatch[1].trim()
    }
    
    // author
    const authorMatch = infoBlock.match(/^\s+author:\s*(.+)$/m)
    if (authorMatch) {
      result.author = authorMatch[1].trim()
    }
    
    // severity
    const severityMatch = infoBlock.match(/^\s+severity:\s*(.+)$/m)
    if (severityMatch) {
      const severity = severityMatch[1].trim().toLowerCase()
      if (['critical', 'high', 'medium', 'low', 'info'].includes(severity)) {
        result.severity = severity
      }
    }
    
    // description (支持多行)
    const descMatch = infoBlock.match(/^\s+description:\s*\|?\s*\n?((?:\s{4,}.+\n?)*)/m)
    if (descMatch && descMatch[1]) {
      result.description = descMatch[1].trim()
    } else {
      const descSimpleMatch = infoBlock.match(/^\s+description:\s*(.+)$/m)
      if (descSimpleMatch) {
        result.description = descSimpleMatch[1].trim()
      }
    }
    
    // tags
    const tagsMatch = infoBlock.match(/^\s+tags:\s*(.+)$/m)
    if (tagsMatch) {
      const tagsStr = tagsMatch[1].trim()
      result.tags = tagsStr.split(',').map(t => t.trim()).filter(t => t)
    }
  }
  
  // 如果没有解析到必要字段，返回null
  if (!result.templateId && !result.name) {
    return null
  }
  
  // 如果没有name，使用templateId
  if (!result.name) {
    result.name = result.templateId
  }
  
  return result
}

// ==================== XRAY POC 转 Nuclei POC ====================

// 解析 XRAY POC 并转换为 Nuclei 格式
function parseXrayToNuclei(xrayContent) {
  try {
    // 解析 XRAY POC 的基本信息
    const xrayPoc = parseXrayPoc(xrayContent)
    if (!xrayPoc) {
      console.warn('Failed to parse XRAY POC')
      return null
    }
    
    // 转换为 Nuclei 格式
    const nucleiContent = convertToNucleiFormat(xrayPoc)
    
    // 从 vulnerability.level 获取严重级别
    const severity = xrayPoc.detail?.vulnerability?.level || xrayPoc.detail?.severity || 'medium'
    
    return {
      templateId: xrayPoc.name || 'xray-converted-poc',
      name: xrayPoc.detail?.name || xrayPoc.name || 'Converted POC',
      author: xrayPoc.detail?.author || 'xray-converter',
      severity: mapXraySeverity(severity),
      description: xrayPoc.detail?.description || '',
      tags: extractXrayTags(xrayPoc),
      content: nucleiContent
    }
  } catch (e) {
    console.error('XRAY to Nuclei conversion error:', e)
    return null
  }
}

// 解析 XRAY POC 结构
function parseXrayPoc(content) {
  const poc = {
    name: '',
    transport: 'http',
    set: {},
    rules: {},
    expression: '',
    detail: {
      author: '',
      links: [],
      vulnerability: { id: '', level: '' }
    }
  }
  
  // 使用 js-yaml 解析更可靠
  try {
    const yaml = jsYaml.load(content)
    if (!yaml) return null
    
    poc.name = yaml.name || ''
    poc.transport = yaml.transport || 'http'
    
    // 解析 set 变量
    if (yaml.set) {
      poc.set = yaml.set
    }
    
    // 解析 detail
    if (yaml.detail) {
      poc.detail.author = yaml.detail.author || ''
      poc.detail.links = yaml.detail.links || []
      if (yaml.detail.vulnerability) {
        poc.detail.vulnerability = {
          id: yaml.detail.vulnerability.id || '',
          level: yaml.detail.vulnerability.level || 'medium'
        }
      }
    }
    
    // 解析 rules
    if (yaml.rules) {
      for (const [ruleName, ruleData] of Object.entries(yaml.rules)) {
        const rule = {
          request: { method: 'GET', path: '/', headers: {}, body: '' },
          expression: ''
        }
        
        if (ruleData.request) {
          rule.request.method = ruleData.request.method || 'GET'
          rule.request.path = ruleData.request.path || '/'
          rule.request.headers = ruleData.request.headers || {}
          rule.request.body = ruleData.request.body || ''
        }
        
        rule.expression = ruleData.expression || ''
        poc.rules[ruleName] = rule
      }
    }
    
    // 解析顶层 expression
    poc.expression = yaml.expression || ''
    
    if (!poc.name && Object.keys(poc.rules).length === 0) {
      return null
    }
    
    return poc
  } catch (e) {
    console.error('YAML parse error:', e)
    return null
  }
}

// 解析 XRAY rules (保留作为备用，主要使用 YAML 解析)
function parseXrayRules(rulesContent) {
  const rules = {}
  
  // 匹配每个规则块 - XRAY 使用 4 空格缩进
  const ruleMatches = rulesContent.matchAll(/^[ ]{4}(\w+):\s*\n((?:[ ]{6,}.+\n?)+)/gm)
  
  for (const match of ruleMatches) {
    const ruleName = match[1]
    const ruleContent = match[2]
    
    const rule = {
      request: { method: 'GET', path: '/', headers: {}, body: '' },
      expression: ''
    }
    
    // 解析 request
    const requestMatch = ruleContent.match(/request:\s*\n((?:\s+.+\n?)+?)(?=\s+expression:|$)/m)
    if (requestMatch) {
      const reqBlock = requestMatch[1]
      
      const methodMatch = reqBlock.match(/method:\s*(.+)/m)
      if (methodMatch) rule.request.method = methodMatch[1].trim()
      
      const pathMatch = reqBlock.match(/path:\s*(.+)/m)
      if (pathMatch) rule.request.path = pathMatch[1].trim().replace(/^["']|["']$/g, '')
      
      const headersMatch = reqBlock.match(/headers:\s*\n((?:\s+.+:\s*.+\n?)+)/m)
      if (headersMatch) {
        const headerLines = headersMatch[1].match(/^\s+(.+?):\s*(.+)$/gm)
        if (headerLines) {
          for (const line of headerLines) {
            const hMatch = line.match(/^\s+(.+?):\s*(.+)$/)
            if (hMatch) {
              rule.request.headers[hMatch[1].trim()] = hMatch[2].trim().replace(/^["']|["']$/g, '')
            }
          }
        }
      }
      
      const bodyMatch = reqBlock.match(/body:\s*\|?\s*\n?([\s\S]*?)(?=\n\s+\w+:|$)/m)
      if (bodyMatch) {
        rule.request.body = bodyMatch[1].trim().replace(/^["']|["']$/g, '')
      } else {
        const simpleBodyMatch = reqBlock.match(/body:\s*(.+)$/m)
        if (simpleBodyMatch) rule.request.body = simpleBodyMatch[1].trim().replace(/^["']|["']$/g, '')
      }
    }
    
    // 解析 expression
    const exprMatch = ruleContent.match(/expression:\s*(.+)/m)
    if (exprMatch) rule.expression = exprMatch[1].trim()
    
    rules[ruleName] = rule
  }
  
  return rules
}

// 转换为 Nuclei 格式
function convertToNucleiFormat(xrayPoc) {
  const lines = []
  
  // ID
  lines.push(`id: ${xrayPoc.name || 'converted-poc'}`)
  lines.push('')
  
  // Info 块
  lines.push('info:')
  lines.push(`  name: ${xrayPoc.detail?.name || xrayPoc.name || 'Converted POC'}`)
  lines.push(`  author: ${xrayPoc.detail?.author || 'xray-converter'}`)
  
  // 从 vulnerability.level 获取严重级别
  const severity = xrayPoc.detail?.vulnerability?.level || xrayPoc.detail?.severity || 'medium'
  lines.push(`  severity: ${mapXraySeverity(severity)}`)
  
  if (xrayPoc.detail?.description) {
    lines.push(`  description: ${xrayPoc.detail.description}`)
  }
  
  // Tags
  const tags = extractXrayTags(xrayPoc)
  if (tags.length > 0) {
    lines.push(`  tags: ${tags.join(',')}`)
  }
  
  // Reference links
  if (xrayPoc.detail?.links && xrayPoc.detail.links.length > 0) {
    lines.push('  reference:')
    for (const link of xrayPoc.detail.links) {
      lines.push(`    - ${link}`)
    }
  }
  
  lines.push('')
  
  // HTTP requests
  const ruleNames = Object.keys(xrayPoc.rules)
  if (ruleNames.length > 0) {
    lines.push('http:')
    
    for (const ruleName of ruleNames) {
      const rule = xrayPoc.rules[ruleName]
      
      // 使用 raw 请求格式
      lines.push('  - raw:')
      lines.push('      - |')
      
      // 构建 raw HTTP 请求
      const method = rule.request.method || 'GET'
      const path = convertXrayPath(rule.request.path, xrayPoc.set)
      lines.push(`        ${method} ${path} HTTP/1.1`)
      lines.push('        Host: {{Hostname}}')
      
      // Headers
      for (const [key, value] of Object.entries(rule.request.headers || {})) {
        if (key.toLowerCase() !== 'host') {
          const convertedValue = convertXrayVariables(value, xrayPoc.set)
          lines.push(`        ${key}: ${convertedValue}`)
        }
      }
      
      // Body
      if (rule.request.body) {
        lines.push('')
        const bodyContent = convertXrayVariables(rule.request.body, xrayPoc.set)
        // 处理多行 body
        const bodyLines = bodyContent.split('\n')
        for (const bodyLine of bodyLines) {
          lines.push(`        ${bodyLine.trim()}`)
        }
      }
      
      lines.push('')
      
      // Matchers
      const matchers = convertXrayExpression(rule.expression)
      if (matchers.length > 0) {
        lines.push('    matchers-condition: and')
        lines.push('    matchers:')
        for (const matcher of matchers) {
          lines.push(`      - type: ${matcher.type}`)
          if (matcher.part) {
            lines.push(`        part: ${matcher.part}`)
          }
          if (matcher.words) {
            lines.push('        words:')
            for (const word of matcher.words) {
              lines.push(`          - "${word}"`)
            }
          }
          if (matcher.regex) {
            lines.push('        regex:')
            for (const r of matcher.regex) {
              lines.push(`          - "${r}"`)
            }
          }
          if (matcher.status) {
            lines.push('        status:')
            for (const s of matcher.status) {
              lines.push(`          - ${s}`)
            }
          }
        }
      }
    }
  }
  
  return lines.join('\n')
}

// 转换 XRAY 路径变量
function convertXrayPath(path, setVars) {
  if (!path) return '/'
  
  let result = path
  // 替换 {{变量}} 为 Nuclei 格式
  result = result.replace(/\{\{(\w+)\}\}/g, (match, varName) => {
    if (setVars && setVars[varName]) {
      // 如果是 randomInt 等函数，转换为 Nuclei 格式
      const value = setVars[varName]
      if (value.includes('randomInt')) {
        return '{{rand_int(1000, 9999)}}'
      }
      if (value.includes('randomLowercase')) {
        return '{{rand_base(8)}}'
      }
      return `{{${varName}}}`
    }
    return match
  })
  
  return result
}

// 转换 XRAY 变量
function convertXrayVariables(str, setVars) {
  if (!str) return str
  
  let result = str
  result = result.replace(/\{\{(\w+)\}\}/g, (match, varName) => {
    if (setVars && setVars[varName]) {
      const value = setVars[varName]
      if (value.includes('randomInt')) {
        return '{{rand_int(1000, 9999)}}'
      }
      if (value.includes('randomLowercase')) {
        return '{{rand_base(8)}}'
      }
    }
    return match
  })
  
  return result
}

// 转换 XRAY expression 为 Nuclei matchers
function convertXrayExpression(expression) {
  const matchers = []
  
  if (!expression) return matchers
  
  // 解析 response.status == xxx
  const statusMatch = expression.match(/response\.status\s*==\s*(\d+)/g)
  if (statusMatch) {
    const statuses = statusMatch.map(m => {
      const match = m.match(/(\d+)/)
      return match ? parseInt(match[1]) : null
    }).filter(s => s !== null)
    
    if (statuses.length > 0) {
      matchers.push({
        type: 'status',
        status: statuses
      })
    }
  }
  
  // 解析 response.body.bcontains(b"xxx") 或 response.body.contains("xxx")
  const bodyContainsMatch = expression.match(/response\.body\.b?contains\s*\(\s*b?["']([^"']+)["']\s*\)/g)
  if (bodyContainsMatch) {
    const words = bodyContainsMatch.map(m => {
      const match = m.match(/\(\s*b?["']([^"']+)["']\s*\)/)
      return match ? match[1] : null
    }).filter(w => w !== null)
    
    if (words.length > 0) {
      matchers.push({
        type: 'word',
        part: 'body',
        words: words
      })
    }
  }
  
  // 解析 response.headers["xxx"].contains("xxx") 或 response.headers[xxx].startsWith(xxx)
  const headerMatch = expression.match(/response\.headers\[["']?(\w+)["']?\]\.(contains|startsWith)\s*\(\s*["']?([^)"']+)["']?\s*\)/g)
  if (headerMatch) {
    const words = headerMatch.map(m => {
      const match = m.match(/\.(contains|startsWith)\s*\(\s*["']?([^)"']+)["']?\s*\)/)
      return match ? match[2] : null
    }).filter(w => w !== null && !w.match(/^\w+$/)) // 排除变量名
    
    if (words.length > 0) {
      matchers.push({
        type: 'word',
        part: 'header',
        words: words
      })
    }
  }
  
  // 解析 response.body.bmatches(xxx) 或 "xxx".bmatches(response.body) - 正则匹配
  const regexMatch = expression.match(/(?:response\.body\.b?matches\s*\(\s*["']([^"']+)["']\s*\)|["']([^"']+)["']\.b?matches\s*\(\s*response\.body\s*\))/g)
  if (regexMatch) {
    const regexes = regexMatch.map(m => {
      const match = m.match(/["']([^"']+)["']/)
      return match ? match[1] : null
    }).filter(r => r !== null)
    
    if (regexes.length > 0) {
      matchers.push({
        type: 'regex',
        part: 'body',
        regex: regexes
      })
    }
  }
  
  // 如果没有解析到任何 matcher，添加一个默认的
  if (matchers.length === 0) {
    matchers.push({
      type: 'status',
      status: [200]
    })
  }
  
  return matchers
}

// 映射 XRAY 严重级别到 Nuclei
function mapXraySeverity(xraySeverity) {
  if (!xraySeverity) return 'medium'
  
  const severityMap = {
    'critical': 'critical',
    'high': 'high',
    'medium': 'medium',
    'low': 'low',
    'info': 'info',
    'informational': 'info'
  }
  
  return severityMap[xraySeverity.toLowerCase()] || 'medium'
}

// 从 XRAY POC 提取标签
function extractXrayTags(xrayPoc) {
  const tags = ['xray-converted']
  
  // 从 name 中提取可能的标签
  if (xrayPoc.name) {
    const name = xrayPoc.name.toLowerCase()
    
    // 常见漏洞类型
    if (name.includes('rce') || name.includes('command')) tags.push('rce')
    if (name.includes('sqli') || name.includes('sql-injection') || name.includes('sql_injection')) tags.push('sqli')
    if (name.includes('xss')) tags.push('xss')
    if (name.includes('ssrf')) tags.push('ssrf')
    if (name.includes('lfi') || name.includes('file-read') || name.includes('readfile')) tags.push('lfi')
    if (name.includes('rfi')) tags.push('rfi')
    if (name.includes('xxe')) tags.push('xxe')
    if (name.includes('upload') || name.includes('writefile')) tags.push('fileupload')
    if (name.includes('unauth') || name.includes('bypass') || name.includes('unauthorized')) tags.push('unauth')
    if (name.includes('disclosure') || name.includes('leak')) tags.push('exposure')
    if (name.includes('deserialization')) tags.push('deserialization')
    if (name.includes('directory_traversal') || name.includes('path-traversal')) tags.push('traversal')
    
    // CVE
    const cveMatch = name.match(/cve-\d{4}-\d+/i)
    if (cveMatch) tags.push(cveMatch[0].toLowerCase())
    
    // CNVD
    const cnvdMatch = name.match(/cnvd-\d{4}-\d+/i)
    if (cnvdMatch) tags.push(cnvdMatch[0].toLowerCase())
  }
  
  // 从 vulnerability id 提取
  if (xrayPoc.detail?.vulnerability?.id) {
    const vulnId = xrayPoc.detail.vulnerability.id
    // CT-xxx 格式
    if (vulnId.startsWith('CT-')) {
      tags.push(vulnId.toLowerCase())
    }
    // CVE 格式
    if (vulnId.toLowerCase().startsWith('cve-')) {
      tags.push(vulnId.toLowerCase())
    }
  }
  
  return [...new Set(tags)] // 去重
}

// 预览转换后的 POC
function previewConvertedPoc(poc) {
  convertedPocPreviewContent.value = poc.content
  convertedPocPreviewVisible.value = true
}

// 复制转换后的 POC
function copyConvertedPoc() {
  navigator.clipboard.writeText(convertedPocPreviewContent.value).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

// 处理文件上传
function handleImportFileChange(uploadFile, uploadFiles) {
  // uploadFile 是当前变化的文件，uploadFiles 是所有文件列表
  console.log('File change:', uploadFile?.name, 'Total files:', uploadFiles?.length)
  
  if (!uploadFile || !uploadFile.raw) {
    console.warn('No file or raw data')
    return
  }
  
  // 检查文件类型
  const fileName = uploadFile.name.toLowerCase()
  if (!fileName.endsWith('.yaml') && !fileName.endsWith('.yml')) {
    ElMessage.warning(`文件 ${uploadFile.name} 不是YAML文件，已跳过`)
    return
  }
  
  // 检查文件是否已处理过（通过标记）
  if (uploadFile._processed) {
    return
  }
  uploadFile._processed = true
  uploadedFileCount.value++
  
  const reader = new FileReader()
  reader.onload = (e) => {
    const content = e.target.result
    console.log('File content loaded:', uploadFile.name, 'Length:', content?.length)
    
    if (!content || content.trim().length === 0) {
      ElMessage.warning(`文件 ${uploadFile.name} 内容为空`)
      return
    }
    
    let parsed
    if (importPocFormat.value === 'xray') {
      parsed = parseXrayToNuclei(content)
    } else {
      parsed = parseYamlToPreview(content)
    }
    
    if (parsed) {
      // 检查是否已存在相同templateId或相同内容
      const existsByTemplateId = importPocPreviews.value.some(p => p.templateId === parsed.templateId)
      const existsByContent = importPocPreviews.value.some(p => p.content.trim() === parsed.content.trim())
      
      if (!existsByTemplateId && !existsByContent) {
        importPocPreviews.value.push(parsed)
        console.log('Added POC:', parsed.templateId)
      } else {
        console.log(`模板 ${parsed.templateId} 已存在（ID或内容重复），已跳过`)
      }
    } else {
      ElMessage.warning(`文件 ${uploadFile.name} 解析失败，请检查格式`)
    }
  }
  reader.onerror = (err) => {
    console.error('File read error:', err)
    ElMessage.error(`文件 ${uploadFile.name} 读取失败`)
  }
  reader.readAsText(uploadFile.raw)
}

// 处理文件移除
function handleImportFileRemove(file) {
  // 文件移除时不做处理，预览列表由用户手动管理
}

// 移除预览项
function removeImportPreview(index) {
  importPocPreviews.value.splice(index, 1)
}

// 清空导入内容
function clearImportContent() {
  importPocContent.value = ''
  importPocPreviews.value = []
  uploadedFileCount.value = 0
  if (importPocUploadRef.value) {
    importPocUploadRef.value.clearFiles()
  }
}

// 执行导入
async function handleImportPocs() {
  if (importPocPreviews.value.length === 0) {
    ElMessage.warning('没有可导入的POC')
    return
  }
  
  importPocLoading.value = true
  
  try {
    // 准备批量导入数据
    const pocs = importPocPreviews.value.map(poc => ({
      name: poc.name,
      templateId: poc.templateId,
      severity: poc.severity,
      tags: poc.tags,
      author: poc.author,
      description: poc.description,
      content: poc.content,
      enabled: importPocEnabled.value
    }))
    
    // 尝试批量导入
    const res = await batchImportCustomPoc({ pocs })
    
    if (res.code === 0) {
      const successCount = res.imported || pocs.length
      const failCount = res.failed || 0
      ElMessage.success(`成功导入 ${successCount} 个POC${failCount > 0 ? `，${failCount} 个失败` : ''}`)
      importPocDialogVisible.value = false
      loadCustomPocs()
    } else {
      // 批量导入失败，回退到逐个导入
      console.warn('批量导入API失败，回退到逐个导入:', res.msg)
      await fallbackSingleImport()
    }
  } catch (e) {
    // API不存在或出错，回退到逐个导入
    console.warn('批量导入API异常，回退到逐个导入:', e)
    await fallbackSingleImport()
  } finally {
    importPocLoading.value = false
  }
}

// 回退到逐个导入
async function fallbackSingleImport() {
  let successCount = 0
  let failCount = 0
  
  for (const poc of importPocPreviews.value) {
    const submitData = {
      name: poc.name,
      templateId: poc.templateId,
      severity: poc.severity,
      tags: poc.tags,
      author: poc.author,
      description: poc.description,
      content: poc.content,
      enabled: importPocEnabled.value
    }
    
    try {
      const res = await saveCustomPoc(submitData)
      if (res.code === 0) {
        successCount++
      } else {
        failCount++
        console.error(`导入 ${poc.templateId} 失败:`, res.msg)
      }
    } catch (e) {
      failCount++
      console.error(`导入 ${poc.templateId} 失败:`, e)
    }
  }
  
  if (successCount > 0) {
    ElMessage.success(`成功导入 ${successCount} 个POC${failCount > 0 ? `，${failCount} 个失败` : ''}`)
    importPocDialogVisible.value = false
    loadCustomPocs()
  } else {
    ElMessage.error('导入失败')
  }
}

// ==================== 导入POC相关函数结束 ====================

function getSeverityType(severity) {
  const map = {
    critical: 'danger',
    high: 'warning',
    medium: '',
    low: 'info',
    info: 'success'
  }
  return map[severity] || 'info'
}

function getNucleiTemplate() {
  return `id: custom-poc-template

info:
  name: Custom POC Template
  author: your-name
  severity: medium
  description: Description of the vulnerability
  tags: custom,poc

http:
  - method: GET
    path:
      - "{{BaseURL}}/vulnerable-path"

    matchers-condition: and
    matchers:
      - type: status
        status:
          - 200

      - type: word
        words:
          - "vulnerable-keyword"
        part: body
`
}

// 解析YAML内容，提取字段
function parseYamlContent() {
  const content = customPocForm.content
  if (!content) return

  // 解析 id 字段
  const idMatch = content.match(/^id:\s*(.+)$/m)
  if (idMatch) {
    customPocForm.templateId = idMatch[1].trim()
  }

  // 解析 info 块中的字段
  const infoMatch = content.match(/info:\s*\n((?:\s+.+\n?)+)/m)
  if (infoMatch) {
    const infoBlock = infoMatch[1]
    
    // name
    const nameMatch = infoBlock.match(/^\s+name:\s*(.+)$/m)
    if (nameMatch) {
      customPocForm.name = nameMatch[1].trim()
    }
    
    // author
    const authorMatch = infoBlock.match(/^\s+author:\s*(.+)$/m)
    if (authorMatch) {
      customPocForm.author = authorMatch[1].trim()
    }
    
    // severity
    const severityMatch = infoBlock.match(/^\s+severity:\s*(.+)$/m)
    if (severityMatch) {
      const severity = severityMatch[1].trim().toLowerCase()
      if (['critical', 'high', 'medium', 'low', 'info'].includes(severity)) {
        customPocForm.severity = severity
      }
    }
    
    // description
    const descMatch = infoBlock.match(/^\s+description:\s*(.+)$/m)
    if (descMatch) {
      customPocForm.description = descMatch[1].trim()
    }
    
    // tags (可能是逗号分隔或YAML数组)
    const tagsMatch = infoBlock.match(/^\s+tags:\s*(.+)$/m)
    if (tagsMatch) {
      const tagsStr = tagsMatch[1].trim()
      // 处理逗号分隔的标签
      const tags = tagsStr.split(',').map(t => t.trim()).filter(t => t)
      customPocForm.tags = tags
      customPocForm.tagsInput = tags.join(', ') // 同步更新输入框
    }
  }
}

// 显示POC验证对话框
function showPocValidateDialog(row) {
  validatePoc.value = row
  pocValidateUrl.value = ''
  pocValidateResult.value = null
  pocValidateLogs.value = []
  currentTaskId = null
  pocValidateDialogVisible.value = true
}

// 轮询定时器
let pollTimer = null

// 清理轮询定时器和日志流
onUnmounted(() => {
  cleanupValidation()
})

// 清理验证相关资源
function cleanupValidation() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  if (logEventSource) {
    logEventSource.close()
    logEventSource = null
  }
  currentTaskId = null
}

// 对话框关闭时清理
function handleValidateDialogClose() {
  cleanupValidation()
  pocValidateLoading.value = false
}

// 开始监听日志流
function startLogStream(taskId) {
  // 关闭之前的连接
  if (logEventSource) {
    logEventSource.close()
  }
  
  pocValidateLogs.value = []
  currentTaskId = taskId
  
  // 连接SSE日志流
  const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
  logEventSource = new EventSource(`${baseUrl}/api/v1/worker/logs/stream`)
  
  logEventSource.onmessage = (event) => {
    try {
      const log = JSON.parse(event.data)
      // 只显示包含当前taskId的日志
      if (log.message && log.message.includes(taskId)) {
        // 提取日志中的关键信息，去掉taskId前缀
        let displayMsg = log.message
        const taskIdPrefix = `[${taskId}] `
        if (displayMsg.startsWith(taskIdPrefix)) {
          displayMsg = displayMsg.substring(taskIdPrefix.length)
        }
        
        pocValidateLogs.value.push({
          level: log.level || 'INFO',
          message: displayMsg,
          timestamp: log.timestamp || new Date().toLocaleTimeString()
        })
        // 限制日志数量
        if (pocValidateLogs.value.length > 50) {
          pocValidateLogs.value.shift()
        }
        // 滚动到底部
        scrollLogsToBottom()
      }
    } catch (e) {
      // 忽略解析错误
    }
  }
  
  logEventSource.onerror = () => {
    // 连接错误时不做处理，轮询会继续
  }
}

// 滚动日志到底部
function scrollLogsToBottom() {
  setTimeout(() => {
    if (logsContainerRef.value) {
      logsContainerRef.value.scrollTop = logsContainerRef.value.scrollHeight
    }
  }, 50)
}

// 执行POC验证
async function handleValidatePoc() {
  if (!pocValidateUrl.value) {
    ElMessage.warning('请输入目标URL')
    return
  }

  pocValidateLoading.value = true
  pocValidateResult.value = null
  pocValidateLogs.value = []

  // 清理之前的轮询和日志流
  cleanupValidation()

  // 添加初始日志
  pocValidateLogs.value.push({
    level: 'INFO',
    message: '正在提交验证任务...',
    timestamp: new Date().toLocaleTimeString()
  })

  try {
    const res = await validatePocApi({
      id: validatePoc.value.id,
      url: pocValidateUrl.value,
      pocType: validatePoc.value.pocType || 'custom'
    })

    if (res.code === 0) {
      pocValidateLogs.value.push({
        level: 'INFO',
        message: `任务已下发，TaskId: ${res.taskId}`,
        timestamp: new Date().toLocaleTimeString()
      })

      // 如果返回了taskId，开始监听日志和轮询结果
      if (res.taskId) {
        startLogStream(res.taskId)
        startPollingResult(res.taskId)
      }
    } else {
      pocValidateLogs.value.push({
        level: 'ERROR',
        message: res.msg || '验证失败',
        timestamp: new Date().toLocaleTimeString()
      })
      ElMessage.error(res.msg || '验证失败')
      pocValidateLoading.value = false
    }
  } catch (e) {
    pocValidateLogs.value.push({
      level: 'ERROR',
      message: '验证请求失败: ' + e.message,
      timestamp: new Date().toLocaleTimeString()
    })
    ElMessage.error('验证请求失败: ' + e.message)
    pocValidateLoading.value = false
  }
}

// 开始轮询查询结果
function startPollingResult(taskId) {
  let pollCount = 0
  const maxPollCount = 60 // 最多轮询60次（约2分钟）

  pollTimer = setInterval(async () => {
    pollCount++
    
    if (pollCount > maxPollCount) {
      clearInterval(pollTimer)
      pollTimer = null
      if (logEventSource) {
        logEventSource.close()
        logEventSource = null
      }
      pocValidateLoading.value = false
      pocValidateLogs.value.push({
        level: 'ERROR',
        message: '验证超时，请检查Worker状态',
        timestamp: new Date().toLocaleTimeString()
      })
      pocValidateResult.value = {
        matched: false,
        severity: validatePoc.value.severity,
        details: '验证超时，请稍后重试或检查Worker状态',
        status: 'TIMEOUT'
      }
      return
    }

    try {
      const res = await getPocValidationResult({ taskId })
      
      if (res.code === 0) {
        // 更新状态显示
        if (res.status === 'SUCCESS' || res.status === 'FAILURE') {
          // 任务完成
          clearInterval(pollTimer)
          pollTimer = null
          if (logEventSource) {
            logEventSource.close()
            logEventSource = null
          }
          pocValidateLoading.value = false

          if (res.results && res.results.length > 0) {
            const result = res.results[0]
            
            pocValidateLogs.value.push({
              level: result.matched ? 'INFO' : 'INFO',
              message: result.matched ? `发现漏洞！目标: ${result.matchedUrl}` : '验证完成，未发现漏洞',
              timestamp: new Date().toLocaleTimeString()
            })
            
            pocValidateResult.value = {
              matched: result.matched,
              severity: result.severity || validatePoc.value.severity,
              details: result.matched 
                ? `目标: ${result.matchedUrl}\n详情: ${result.details || '匹配成功'}${result.output ? '\n\n输出:\n' + result.output : ''}`
                : `目标: ${result.matchedUrl}\n${result.details || '未发现漏洞'}`,
              status: res.status
            }
            
            if (result.matched) {
              ElMessage.success('发现漏洞！')
            } else {
              ElMessage.info('验证完成，未发现漏洞')
            }
          } else {
            pocValidateLogs.value.push({
              level: 'INFO',
              message: res.status === 'FAILURE' ? '验证失败' : '验证完成',
              timestamp: new Date().toLocaleTimeString()
            })
            pocValidateResult.value = {
              matched: false,
              severity: validatePoc.value.severity,
              details: res.status === 'FAILURE' ? '验证失败' : '验证完成，未发现漏洞',
              status: res.status
            }
          }
        }
      }
    } catch (e) {
      console.error('Poll result error:', e)
    }
  }, 2000) // 每2秒轮询一次
}

// 默认模板选择变化
function handleTemplateSelectionChange(selection) {
  selectedTemplates.value = selection
}

// 显示单个默认模板验证对话框
function showTemplateValidateDialog(row) {
  validatePoc.value = {
    id: row._id || row.id,
    name: row.name,
    templateId: row.id,
    severity: row.severity,
    pocType: 'nuclei'  // 标记为nuclei默认模板
  }
  pocValidateUrl.value = ''
  pocValidateResult.value = null
  pocValidateLogs.value = []
  currentTaskId = null
  pocValidateDialogVisible.value = true
}

// 显示默认模板批量验证对话框
function showTemplateBatchValidateDialog() {
  if (selectedTemplates.value.length === 0) {
    ElMessage.warning('请先选择要验证的模板')
    return
  }
  templateBatchValidateUrls.value = ''
  batchTargetInputType.value = 'text'
  templateBatchValidateLogs.value = []
  templateBatchValidateResults.value = []
  templateBatchValidateProgress.total = 0
  templateBatchValidateProgress.completed = 0
  currentBatchId = null
  templateBatchValidateDialogVisible.value = true
}

// 处理批量URL文件上传
function handleBatchUrlFileChange(file) {
  const reader = new FileReader()
  reader.onload = (e) => {
    templateBatchValidateUrls.value = e.target.result
  }
  reader.readAsText(file.raw)
}

// 处理批量URL文件移除
function handleBatchUrlFileRemove() {
  templateBatchValidateUrls.value = ''
}

// 导出验证结果
function handleExportResults(command) {
  let dataToExport = templateBatchValidateResults.value
  
  if (command === 'matched') {
    dataToExport = dataToExport.filter(r => r.matched)
  }
  
  if (dataToExport.length === 0) {
    ElMessage.warning('没有可导出的数据')
    return
  }
  
  const timestamp = new Date().toISOString().slice(0, 19).replace(/[:-]/g, '')
  
  if (command === 'csv') {
    // 导出CSV格式
    const headers = ['模板名称', '模板ID', '级别', '结果', '匹配URL', '详情']
    const rows = dataToExport.map(r => [
      r.pocName || '',
      r.templateId || r.pocId || '',
      r.severity || '',
      r.matched ? '匹配' : '未匹配',
      r.matchedUrl || '',
      (r.details || '').replace(/[\n\r]/g, ' ')
    ])
    
    const csvContent = [headers, ...rows]
      .map(row => row.map(cell => `"${cell}"`).join(','))
      .join('\n')
    
    const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8' })
    downloadFile(blob, `poc_validation_results_${timestamp}.csv`)
  } else {
    // 导出JSON格式
    const jsonContent = JSON.stringify(dataToExport, null, 2)
    const blob = new Blob([jsonContent], { type: 'application/json' })
    downloadFile(blob, `poc_validation_results_${timestamp}.json`)
  }
  
  ElMessage.success('导出成功')
}

// 下载文件
function downloadFile(blob, filename) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

// 清理批量验证资源
function cleanupBatchValidation() {
  if (batchPollTimer) {
    clearInterval(batchPollTimer)
    batchPollTimer = null
  }
  if (batchLogEventSource) {
    batchLogEventSource.close()
    batchLogEventSource = null
  }
  currentBatchId = null
  currentBatchTaskIds.value = []
}

// 批量验证对话框关闭
function handleBatchValidateDialogClose() {
  cleanupBatchValidation()
  templateBatchValidateLoading.value = false
}

// 开始批量验证日志流
function startBatchLogStream(batchId) {
  if (batchLogEventSource) {
    batchLogEventSource.close()
  }
  
  currentBatchId = batchId
  
  const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
  batchLogEventSource = new EventSource(`${baseUrl}/api/v1/worker/logs/stream`)
  
  batchLogEventSource.onmessage = (event) => {
    try {
      const log = JSON.parse(event.data)
      // 只显示当前批次任务的日志
      if (log.message && currentBatchTaskIds.value.length > 0) {
        // 检查日志是否属于当前批次的任务
        const isCurrentBatchLog = currentBatchTaskIds.value.some(taskId => log.message.includes(taskId))
        if (!isCurrentBatchLog) return
        
        let displayMsg = log.message
        // 去掉taskId前缀
        const taskIdMatch = displayMsg.match(/^\[poc-validate-\d+\]\s*/)
        if (taskIdMatch) {
          displayMsg = displayMsg.substring(taskIdMatch[0].length)
        }
        
        templateBatchValidateLogs.value.push({
          level: log.level || 'INFO',
          message: displayMsg,
          timestamp: log.timestamp || new Date().toLocaleTimeString()
        })
        
        if (templateBatchValidateLogs.value.length > 100) {
          templateBatchValidateLogs.value.shift()
        }
        
        // 滚动到底部
        setTimeout(() => {
          if (batchLogsContainerRef.value) {
            batchLogsContainerRef.value.scrollTop = batchLogsContainerRef.value.scrollHeight
          }
        }, 50)
      }
    } catch (e) {
      // 忽略解析错误
    }
  }
}

// 执行默认模板批量验证
async function handleTemplateBatchValidate() {
  const urls = batchTargetUrls.value
  if (urls.length === 0) {
    ElMessage.warning('请输入有效的目标URL')
    return
  }
  
  if (selectedTemplates.value.length === 0) {
    ElMessage.warning('请先选择要验证的模板')
    return
  }

  templateBatchValidateLoading.value = true
  templateBatchValidateLogs.value = []
  templateBatchValidateResults.value = []
  
  // 总任务数 = 模板数 × URL数
  const totalTasks = selectedTemplates.value.length * urls.length
  templateBatchValidateProgress.total = totalTasks
  templateBatchValidateProgress.completed = 0
  
  cleanupBatchValidation()

  templateBatchValidateLogs.value.push({
    level: 'INFO',
    message: `正在提交批量验证任务，${selectedTemplates.value.length} 个模板 × ${urls.length} 个目标 = ${totalTasks} 个任务...`,
    timestamp: new Date().toLocaleTimeString()
  })

  // 为每个模板和URL组合创建验证任务
  const taskIds = []
  const batchId = `batch-${Date.now()}`
  currentBatchId = batchId
  currentBatchTaskIds.value = [] // 清空之前的任务ID

  for (const url of urls) {
    for (const tpl of selectedTemplates.value) {
      try {
        const res = await validatePocApi({
          id: tpl._id || tpl.id,
          url: url,
          pocType: 'nuclei'
        })

        if (res.code === 0 && res.taskId) {
          taskIds.push(res.taskId)
        } else {
          templateBatchValidateLogs.value.push({
            level: 'ERROR',
            message: `${tpl.name || tpl.id} -> ${url} 下发失败: ${res.msg}`,
            timestamp: new Date().toLocaleTimeString()
          })
        }
      } catch (e) {
        templateBatchValidateLogs.value.push({
          level: 'ERROR',
          message: `${tpl.name || tpl.id} -> ${url} 下发失败: ${e.message}`,
          timestamp: new Date().toLocaleTimeString()
        })
      }
    }
  }

  if (taskIds.length > 0) {
    // 保存任务ID列表并启动日志流
    currentBatchTaskIds.value = taskIds
    startBatchLogStream(batchId)
    
    templateBatchValidateLogs.value.push({
      level: 'INFO',
      message: `共下发 ${taskIds.length} 个任务，开始轮询结果...`,
      timestamp: new Date().toLocaleTimeString()
    })
    startBatchPolling(batchId, taskIds)
  } else {
    templateBatchValidateLoading.value = false
    templateBatchValidateLogs.value.push({
      level: 'ERROR',
      message: '所有任务下发失败',
      timestamp: new Date().toLocaleTimeString()
    })
  }
}

// 批量验证轮询
function startBatchPolling(batchId, taskIds) {
  let pollCount = 0
  const maxPollCount = 120 // 最多轮询120次（约4分钟）
  const completedTasks = new Set()

  batchPollTimer = setInterval(async () => {
    pollCount++
    
    if (pollCount > maxPollCount) {
      cleanupBatchValidation()
      templateBatchValidateLoading.value = false
      templateBatchValidateLogs.value.push({
        level: 'ERROR',
        message: '批量验证超时',
        timestamp: new Date().toLocaleTimeString()
      })
      return
    }

    // 轮询每个任务的结果
    for (const taskId of taskIds) {
      if (completedTasks.has(taskId)) continue
      
      try {
        const res = await getPocValidationResult({ taskId })
        
        if (res.code === 0 && (res.status === 'SUCCESS' || res.status === 'FAILURE')) {
          completedTasks.add(taskId)
          templateBatchValidateProgress.completed = completedTasks.size
          
          if (res.results && res.results.length > 0) {
            for (const result of res.results) {
              templateBatchValidateResults.value.push(result)
            }
          }
        }
      } catch (e) {
        console.error('Poll batch result error:', e)
      }
    }

    // 检查是否全部完成
    if (completedTasks.size >= taskIds.length) {
      cleanupBatchValidation()
      templateBatchValidateLoading.value = false
      
      const matchedCount = templateBatchValidateResults.value.filter(r => r.matched).length
      templateBatchValidateLogs.value.push({
        level: 'INFO',
        message: `批量验证完成，发现 ${matchedCount} 个漏洞`,
        timestamp: new Date().toLocaleTimeString()
      })
      
      if (matchedCount > 0) {
        ElMessage.success(`批量验证完成，发现 ${matchedCount} 个漏洞`)
      } else {
        ElMessage.info('批量验证完成，未发现漏洞')
      }
    }
  }, 2000)
}
</script>

<style lang="scss" scoped>
.poc-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .tip-text {
    color: var(--el-text-color-secondary);
    font-size: 13px;
    margin-bottom: 15px;
    padding: 10px 12px;
    background: var(--el-fill-color-light);
    border-radius: 4px;
    border-left: 3px solid var(--el-color-primary-light-5);
    line-height: 1.6;
  }

  .filter-form {
    margin-bottom: 15px;
  }

  .stats-bar {
    margin-bottom: 15px;
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
  }

  .pagination {
    margin-top: 20px;
    justify-content: flex-end;
  }

  .template-content-wrapper {
    :deep(.el-textarea__inner) {
      background-color: #1e1e1e;
      color: #d4d4d4;
      border: 1px solid #3c3c3c;
    }
  }

  .yaml-editor-wrapper {
    :deep(.el-textarea__inner) {
      background-color: #1e1e1e;
      color: #d4d4d4;
      border: 1px solid #3c3c3c;
      font-family: 'Consolas', 'Monaco', monospace;
      font-size: 13px;
    }
  }

  .validate-logs {
    margin-top: 15px;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    overflow: hidden;

    .logs-header {
      padding: 8px 15px;
      background: #f5f7fa;
      border-bottom: 1px solid #dcdfe6;
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 500;
    }

    .logs-content {
      background: #1e1e1e;
      color: #d4d4d4;
      font-family: 'Consolas', 'Monaco', monospace;
      font-size: 12px;
      padding: 10px;
      max-height: 200px;
      overflow-y: auto;

      .log-line {
        padding: 2px 0;
        line-height: 1.5;

        .log-time {
          color: #6a9955;
          margin-right: 8px;
        }

        .log-level {
          margin-right: 8px;
          font-weight: bold;
        }

        &.log-info .log-level {
          color: #4fc3f7;
        }

        &.log-error .log-level {
          color: #f44336;
        }

        &.log-warn .log-level {
          color: #ff9800;
        }

        .log-msg {
          color: #d4d4d4;
        }
      }
    }
  }

  .validate-result {
    margin-top: 15px;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    overflow: hidden;

    .result-header {
      padding: 10px 15px;
      background: #f5f7fa;
      border-bottom: 1px solid #dcdfe6;
    }

    .result-details {
      margin: 0;
      padding: 12px 15px;
      font-family: 'Consolas', 'Monaco', monospace;
      font-size: 13px;
      background: #1e1e1e;
      color: #d4d4d4;
      white-space: pre-wrap;
      word-break: break-all;
      max-height: 300px;
      overflow-y: auto;
    }
  }

  .selected-templates {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
  }

  .batch-validate-progress {
    margin-top: 15px;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    overflow: hidden;

    .progress-header {
      padding: 10px 15px;
      background: #f5f7fa;
      border-bottom: 1px solid #dcdfe6;
      display: flex;
      align-items: center;
    }

    .logs-content {
      background: #1e1e1e;
      color: #d4d4d4;
      font-family: 'Consolas', 'Monaco', monospace;
      font-size: 12px;
      padding: 10px;
      overflow-y: auto;

      .log-line {
        padding: 2px 0;
        line-height: 1.5;

        .log-time {
          color: #6a9955;
          margin-right: 8px;
        }

        .log-level {
          margin-right: 8px;
          font-weight: bold;
        }

        &.log-info .log-level {
          color: #4fc3f7;
        }

        &.log-error .log-level {
          color: #f44336;
        }

        .log-msg {
          color: #d4d4d4;
        }
      }
    }
  }

  .batch-validate-results {
    margin-top: 15px;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    overflow: hidden;

    .results-header {
      padding: 10px 15px;
      background: #f5f7fa;
      border-bottom: 1px solid #dcdfe6;
      display: flex;
      align-items: center;
      font-weight: 500;
    }
  }

  .import-preview {
    margin-top: 15px;
    border: 1px solid #dcdfe6;
    border-radius: 4px;
    overflow: hidden;

    .preview-header {
      padding: 10px 15px;
      background: #f5f7fa;
      border-bottom: 1px solid #dcdfe6;
      display: flex;
      align-items: center;
      font-weight: 500;
    }
  }
}
</style>
